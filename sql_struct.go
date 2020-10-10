package main

import (
	"bytes"
	"fmt"
	"go/format"
	"regexp"
	"strings"
)

//创建table sql语句
type SqlType struct {
	Tablename string         //数据表名
	Fields    []SqlFieldType //table包含的字段
}

//sql语句 单个字段
type SqlFieldType struct {
	Name     string //字段名
	Type     string //字段类型
	Typemore string //字段类型 补存
	Comment  string //注释
}

//解析创建table sql语句, 支持1个或者多个 创建table sql语句
func Sql2Struct(sqlStr string) {
	var sqlTypes []SqlType

	//
	tablePatternStart := `(?im)^\s*create\s+table\s+(?P<table>\w+)\s*\(?`
	tablePatternEnd := `^\s*\);?`
	regStart := regexp.MustCompile(tablePatternStart)
	regEnd := regexp.MustCompile(tablePatternEnd)

	var findStart, findEnd bool
	var sqlType SqlType
	//解析单个字段
	fieldPattern := `(?m)\s*(?P<name>[\w-_]+)\s+(?P<type>[\w()]+)\s*(?P<typemore>[\w\s]*)\s*(?P<comment>comment\s+'.+')?,`
	for _, s := range strings.Split(sqlStr, "\n") {
		if regStart.MatchString(s) { //找到开头
			m := getParams(tablePatternStart, s)
			sqlType.Tablename = m["table"]

			findStart = true
			findEnd = false
			fmt.Println("start", s)
			continue
		}
		if regEnd.MatchString(s) { //找到结尾
			fmt.Println("end", s)
			e := SqlType{}
			e.Tablename = sqlType.Tablename
			e.Fields = make([]SqlFieldType, 0, len(sqlType.Fields))
			for _, v := range sqlType.Fields {
				e.Fields = append(e.Fields, v)
			}
			sqlType.Tablename = ""
			sqlType.Fields = sqlType.Fields[0:0]

			sqlTypes = append(sqlTypes, e)

			findStart = false
			findEnd = true
			continue
		}

		if findStart && !findEnd { //找字段
			fmt.Println("field", s)
			ms := getAllParams(fieldPattern, s)
			for _, m := range ms {
				var f SqlFieldType
				f.Name, f.Type, f.Typemore, f.Comment = m["name"], m["type"], m["typemore"], m["comment"]
				sqlType.Fields = append(sqlType.Fields, f)
			}
			continue
		}
	}

	//匹配一条create table的sql语句 正则表达式
	//	tablePattern := `(?im)^\s*create\s+table\s+(?P<table>\w+)\s*\((?s)(?P<fields>.+)\)\s*;?`
	//找出所有create table的sql语句
	//	sqls := getAllParams(tablePattern, sqlStr)
	/*
		//保存解析后的多条create table sql语句
		for _, sql := range sqls { //每次循环处理一个sql创建语句
			fmt.Println(sql)
			fmt.Printf("%s\n", strings.Repeat("~", 50))

			var sqlType SqlType
			tablename, fields := sql["table"], sql["fields"]
			sqlType.Tablename = tablename
			//解析单个字段
			fieldPattern := `(?m)\s*(?P<name>[\w-_]+)\s+(?P<type>[\w()]+)\s*(?P<typemore>[\w\s]*)\s*(?P<comment>comment\s+'.+')?,`
			ms := getAllParams(fieldPattern, fields)
			for _, m := range ms {
				var f SqlFieldType
				f.Name, f.Type, f.Typemore, f.Comment = m["name"], m["type"], m["typemore"], m["comment"]
				sqlType.Fields = append(sqlType.Fields, f)
			}
			if len(sqlType.Fields) > 0 {
				sqlTypes = append(sqlTypes, sqlType)
			}
		}
	*/
	Sql2StructOutput(sqlTypes)

}

//按照go语法输出
func Sql2StructOutput(sqlTypes []SqlType) {
	fmt.Printf("%s\n", strings.Repeat("~", 60))
	var buff bytes.Buffer
	for _, v := range sqlTypes {
		buff.Reset()
		//
		fmt.Fprintf(&buff, "type %s struct {//%s\n", HFiledname(v.Tablename), v.Tablename)
		for _, field := range v.Fields {
			fmt.Fprintf(&buff, "%s %s %s //%s\n",
				HFiledname(field.Name), HFiledtype(field.Name, field.Type),
				fmt.Sprintf("`json:\"%s\" xorm:\"%s\"`", field.Name, field.Name),
				HFiledComment(field.Comment),
			)
		}
		fmt.Fprintf(&buff, "}")
		//format go 代码
		result, err := format.Source(buff.Bytes())
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n\n\n", result)
	}
}

//变量名按照驼峰规则转换
func HFiledname(in string) (out string) {
	sli := strings.Split(in, "_")
	for _, s := range sli {
		out += strings.Title(s)
	}
	return
}

//转换字段类型到go类型
func HFiledtype(name, in string) (out string) {
	s := strings.ToLower(in)
	if strings.Contains(s, "int") {
		if strings.Contains(strings.ToLower(name), "id") { //字段中包含id的
			out = "int64"
		} else {
			out = "int"
		}
		return
	}
	if strings.Contains(s, "text") {
		out = "string"
		return
	}
	if strings.Contains(s, "varchar") {
		out = "string"
		return
	}
	if strings.Contains(s, "datetime") {
		out = "time.Time"
		return
	}
	if strings.Contains(s, "date") {
		out = "string"
		return
	}
	if strings.Contains(s, "bool") {
		out = "bool"
		return
	}
	return
}

//处理字段注释
func HFiledComment(in string) (out string) {
	start := strings.IndexByte(in, '\'')
	end := strings.LastIndexByte(in, '\'')

	if start > 0 && end > 0 && end >= start+1 {
		out = in[start+1 : end]
	}
	return
}

//用regEx解析content,返回第一个满足条件的数据
func getParams(regEx, content string) (paramsMap map[string]string) {
	paramsMap = make(map[string]string)

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(content)
	//fmt.Printf("match: %q\n", match)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return
}

//用regEx解析content,返回所有满足条件的数据
func getAllParams(regEx, content string) (paramsMap []map[string]string) {
	paramsMap = make([]map[string]string, 0)

	var compRegEx = regexp.MustCompile(regEx)
	matches := compRegEx.FindAllStringSubmatch(content, -1)
	//fmt.Printf("matches: %q\n", matches)
	for i := range matches {
		_map := make(map[string]string)
		for j, name := range compRegEx.SubexpNames() {
			if j > 0 && j <= len(matches[i]) {
				_map[name] = matches[i][j]
			}
		}
		if len(_map) > 0 {
			paramsMap = append(paramsMap, _map)
		}
	}
	return
}
