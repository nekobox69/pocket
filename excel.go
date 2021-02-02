// Package pocket Create at 2021-01-29 10:17
package pocket

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	excelize "github.com/360EntSecGroup-Skylar/excelize/v2"
) 

const (
	headerStyle = `{
               "fill":{"type":"pattern","color":["#FFC408"],"pattern":1},
               "border":[
                         {"type":"left","color":"000000","style":1},
                         {"type":"right","color":"000000","style":1},
                         {"type":"top","color":"000000","style":1},
                         {"type":"bottom","color":"000000","style":1}
                         ]
               }`
	contentStyle = `{
               "border":[
                         {"type":"left","color":"000000","style":1},
                         {"type":"right","color":"000000","style":1},
                         {"type":"top","color":"000000","style":1},
                         {"type":"bottom","color":"000000","style":1}
                         ],
              "alignment":{
                           "horizontal":"left",
                           "ident":1,
                           "justify_last_line":true,
                           "vertical":"",
                           "wrap_text":true
                           }
              }`
	a = 97
)

// Formatter data formatter
type Formatter struct {
	Enum string `schema:"enum"`
	Time string `schema:"time"`
}

// Format format value
type Format interface {
	format(value interface{}) string
}

// enumFormatter 枚举类型formatter
type enumFormatter struct {
	enum map[string]string
}

func (e enumFormatter) format(value interface{}) string {
	return e.enum[fmt.Sprintf("%v", value)]
}

// timeExportFormatter 时间导出类型formatter
type timeExportFormatter struct {
	timeLayout string
}

func (t timeExportFormatter) format(value interface{}) string {
	val, err := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64)
	if nil != err {
		DefaultLogger.Error(err.Error())
		return ""
	}
	time.Unix(val, 0)
	return time.Unix(val, 0).Format(t.timeLayout)
}

// timeImportFormatter 时间导入类型formatter
type timeImportFormatter struct {
	timeLayout string
}

func (t timeImportFormatter) format(value interface{}) string {
	m, err := time.Parse(t.timeLayout, fmt.Sprintf("%v", value))
	if nil != err {
		DefaultLogger.Error(err.Error())
		return "0"
	}
	return fmt.Sprintf("%d", m.Unix())
}

// Sheet 表
type Sheet struct {
	Name         string            `json:"name"`              // sheet 名称
	T            reflect.Type      `json:"-"`                 // 列的类型
	Result       *[]interface{}    `json:"result,omitempty"`  // 导入结果
	Content      []interface{}     `json:"content,omitempty"` // 导出数据
	HeaderStyle  string            `json:"-"`                 // 头部单元格样式
	ContentStyle string            `json:"-"`                 // 内容单元格样式
	Panes        []string          `json:"-"`
	Columns      map[string]Column `json:"-"` // 内容单元格详细设置
}

// Column 单元格设置
type Column struct {
	cell         string  // 对应的列
	Width        float64 // 单元格宽度
	Merge        bool    // 是否合并
	MergeExclude string  // 合并排除
	Style        string  // 单元格样式，覆盖Sheet的ContentStyle
}

type field struct {
	Idx int
	Tag string
}

type mergeItem struct {
	Col     string
	Start   int
	End     int
	Val     string
	Exclude string
}

// Export export excel
func Export(sheets []Sheet) (*bytes.Buffer, error) {
	xlsx := excelize.NewFile()
	last := 0
	for _, s := range sheets {
		column := make(map[string]Column, 0)
		formatter := make(map[string]Format, 0)
		last = xlsx.NewSheet(s.Name)
		t := s.T
		if nil == t {
			if len(s.Content) == 0 {
				DefaultLogger.Error("无法识别类型")
				return nil, errors.New("无法识别类型")
			}
			t = reflect.TypeOf(s.Content[0])
		}
		if t.Kind() != reflect.Ptr && t.Kind() != reflect.Struct {
			DefaultLogger.Error("不支持的类型，只能是指针或结构体")
			return nil, errors.New("不支持的类型，只能是指针或结构体")
		}
		size := len(s.Content)
		switch t.Kind() {
		case reflect.Ptr:
			style, err := xlsx.NewStyle(s.HeaderStyle)
			if nil != err {
				DefaultLogger.Warn("创建表头样式失败")
			}
			for j := 0; j < t.Elem().NumField(); j++ {
				tag := t.Elem().Field(j).Tag.Get("excel_column")
				if "" != tag {
					if nil == err {
						xlsx.SetCellStyle(s.Name, fmt.Sprintf("%s1", string(a+j)),
							fmt.Sprintf("%s1", string(a+j)), style)
					}
					title := tag
					if nil != s.Columns {
						if c, ok := s.Columns[tag]; ok {
							column[tag] = Column{
								cell:         string(a + j),
								Width:        c.Width,
								Merge:        c.Merge,
								MergeExclude: c.MergeExclude,
								Style:        c.Style,
							}
							xlsx.SetColWidth(s.Name, fmt.Sprintf("%s", string(a+j)),
								fmt.Sprintf("%s", string(a+j)), c.Width)
						} else {
							column[tag] = Column{cell: string(a + j)}
						}
					} else {
						column[tag] = Column{cell: string(a + j)}
					}

					xlsx.SetCellValue(s.Name, fmt.Sprintf("%s1", string(a+j)), title)

					form := t.Elem().Field(j).Tag.Get("excel_formatter")
					if "" != form {
						var f Formatter
						v, err := url.ParseQuery(form)
						if nil != err {
							DefaultLogger.Error(err.Error())
							return nil, err
						}
						err = DecodeQuery(&f, v)
						if nil != err {
							DefaultLogger.Error(err.Error())
							return nil, err
						}
						if "" != f.Enum {
							list := strings.Split(f.Enum, ",")
							items := make([]string, 0)
							if len(list) > 0 {
								enum := enumFormatter{enum: make(map[string]string, 0)}
								for i := 0; i < len(list); i++ {
									enum.enum[list[i][:strings.Index(list[i], ":")]] = list[i][strings.Index(list[i], ":")+1:]
									items = append(items, list[i][strings.Index(list[i], ":")+1:])
								}
								formatter[tag] = enum
							}
							if size > 0 {
								dvRange := excelize.NewDataValidation(true)
								dvRange.Sqref = fmt.Sprintf("%s2:%s%d", string(a+j), string(a+j), size+1)
								err = dvRange.SetDropList(items)
								if nil != err {
									DefaultLogger.Error(err.Error())
									//return nil, err
								}
								err = xlsx.AddDataValidation(s.Name, dvRange)
								if nil != err {
									DefaultLogger.Error(err.Error())
									//return nil, err
								}
							}
						}
						if "" != f.Time {
							formatter[tag] = timeExportFormatter{timeLayout: f.Time}
						}
					}
				}
			}
		case reflect.Struct:
			style, err := xlsx.NewStyle(s.HeaderStyle)
			if nil != err {
				DefaultLogger.Warn("创建表头样式失败")
			}
			for j := 0; j < t.NumField(); j++ {
				tag := t.Field(j).Tag.Get("excel_column")
				if "" != tag {
					if nil == err {
						xlsx.SetCellStyle(s.Name, fmt.Sprintf("%s1", string(97+j)),
							fmt.Sprintf("%s1", string(97+j)), style)
					}
					title := tag
					if nil != s.Columns {
						if c, ok := s.Columns[tag]; ok {
							column[tag] = Column{
								cell:         string(a + j),
								Width:        c.Width,
								Merge:        c.Merge,
								MergeExclude: c.MergeExclude,
								Style:        c.Style,
							}
							xlsx.SetColWidth(s.Name, fmt.Sprintf("%s", string(a+j)),
								fmt.Sprintf("%s", string(a+j)), c.Width)
						} else {
							column[tag] = Column{cell: string(a + j)}
						}
					} else {
						column[tag] = Column{cell: string(a + j)}
					}

					xlsx.SetCellValue(s.Name, fmt.Sprintf("%s1", string(a+j)), title)
					form := t.Field(j).Tag.Get("excel_formatter")
					if "" != form {
						var f Formatter
						v, err := url.ParseQuery(form)
						if nil != err {
							DefaultLogger.Error(err.Error())
							return nil, err
						}
						err = DecodeQuery(&f, v)
						if nil != err {
							DefaultLogger.Error(err.Error())
							return nil, err
						}
						if "" != f.Enum {
							list := strings.Split(f.Enum, ",")
							items := make([]string, 0)
							if len(list) > 0 {
								enum := enumFormatter{enum: make(map[string]string, 0)}
								for i := 0; i < len(list); i++ {
									enum.enum[list[i][:strings.Index(list[i], ":")]] = list[i][strings.Index(list[i], ":")+1:]
									items = append(items, list[i][strings.Index(list[i], ":")+1:])
								}
								formatter[tag] = enum
							}
							if size > 0 {
								dvRange := excelize.NewDataValidation(true)
								dvRange.Sqref = fmt.Sprintf("%s2:%s%d", string(a+j), string(a+j), size+1)
								err = dvRange.SetDropList(items)
								if nil != err {
									DefaultLogger.Error(err.Error())
									//return nil, err
								}
								err = xlsx.AddDataValidation(s.Name, dvRange)
								if nil != err {
									DefaultLogger.Error(err.Error())
									//return nil, err
								}
							}
						}
						if "" != f.Time {
							formatter[tag] = timeExportFormatter{timeLayout: f.Time}
						}
					}
				}
			}
		}
		if size == 0 {
			continue
		}
		style, err := xlsx.NewStyle(s.ContentStyle)
		if nil != err {
			DefaultLogger.Warn("创建表内容样式失败")
		}
		merge := make(map[string]mergeItem)
		for index, r := range s.Content {
			switch t.Kind() {
			case reflect.Ptr:
				for j := 0; j < t.Elem().NumField(); j++ {
					tag := t.Elem().Field(j).Tag.Get("excel_column")
					if c, ok := column[tag]; ok {
						if nil == err {
							xlsx.SetCellStyle(s.Name, fmt.Sprintf("%s%d", c.cell, index+2),
								fmt.Sprintf("%s%d", c.cell, index+2), style)
						}
						if len(c.Style) != 0 {
							columnStyle, err := xlsx.NewStyle(c.Style)
							if nil == err {
								xlsx.SetCellStyle(s.Name, fmt.Sprintf("%s%d", c.cell, index+2),
									fmt.Sprintf("%s%d", c.cell, index+2), columnStyle)
							}
						}
						val := reflect.ValueOf(r).Elem().Field(j).Interface()
						if f, ok := formatter[tag]; ok {
							val = f.format(reflect.ValueOf(r).Elem().Field(j).Interface())
						}
						xlsx.SetCellValue(s.Name, fmt.Sprintf("%s%d", c.cell, index+2), val)
						if c.Merge {
							if m, ok := merge[tag]; ok {
								v := fmt.Sprintf("%v", val)
								if v != m.Val {
									if index-m.Start > -1 && "" != m.Val {
										// 合并列
										if len(m.Exclude) == 0 || !strings.HasPrefix(m.Val, m.Exclude) {
											xlsx.MergeCell(s.Name, fmt.Sprintf("%s%d", m.Col, m.Start),
												fmt.Sprintf("%s%d", m.Col, index+1))
										}
									}
									merge[tag] = mergeItem{
										Col:     c.cell,
										Start:   index + 2,
										End:     0,
										Val:     fmt.Sprintf("%v", val),
										Exclude: c.MergeExclude,
									}
								}
							} else {
								merge[tag] = mergeItem{
									Col:     c.cell,
									Start:   index + 2,
									End:     0,
									Val:     fmt.Sprintf("%v", val),
									Exclude: c.MergeExclude,
								}
							}
						}
					}
				}
			case reflect.Struct:
				for j := 0; j < t.NumField(); j++ {
					tag := t.Field(j).Tag.Get("excel_column")
					if c, ok := column[tag]; ok {
						if nil == err {
							xlsx.SetCellStyle(s.Name, fmt.Sprintf("%s%d", c, index+2),
								fmt.Sprintf("%s%d", c, index+2), style)
						}
						if len(c.Style) != 0 {
							columnStyle, err := xlsx.NewStyle(c.Style)
							if nil == err {
								xlsx.SetCellStyle(s.Name, fmt.Sprintf("%s%d", c.cell, index+2),
									fmt.Sprintf("%s%d", c.cell, index+2), columnStyle)
							}
						}
						val := reflect.ValueOf(r).Field(j).Interface()
						if f, ok := formatter[tag]; ok {
							val = f.format(reflect.ValueOf(r).Field(j).Interface())
						}
						xlsx.SetCellValue(s.Name, fmt.Sprintf("%s%d", c.cell, index+2), val)
						if c.Merge {
							if m, ok := merge[tag]; ok {
								v := fmt.Sprintf("%v", val)
								if v != m.Val {
									if index-m.Start > -1 && "" != m.Val {
										// 合并列
										if len(m.Exclude) == 0 || !strings.HasPrefix(m.Val, m.Exclude) {
											xlsx.MergeCell(s.Name, fmt.Sprintf("%s%d", m.Col, m.Start),
												fmt.Sprintf("%s%d", m.Col, index+1))
										}
									}
									merge[tag] = mergeItem{
										Col:     c.cell,
										Start:   index + 2,
										End:     0,
										Val:     fmt.Sprintf("%v", val),
										Exclude: c.MergeExclude,
									}
								}
							} else {
								merge[tag] = mergeItem{
									Col:     c.cell,
									Start:   index + 2,
									End:     0,
									Val:     fmt.Sprintf("%v", val),
									Exclude: c.MergeExclude,
								}
							}
						}
					}
				}
			}
		}

		for _, v := range merge {
			if size-v.Start > 0 && "" != v.Val {
				// 合并列
				if len(v.Exclude) == 0 || !strings.HasPrefix(v.Val, v.Exclude) {
					xlsx.MergeCell(s.Name, fmt.Sprintf("%s%d", v.Col, v.Start),
						fmt.Sprintf("%s%d", v.Col, size+1))
				}
			}
		}
		for _, p := range s.Panes {
			xlsx.SetPanes(s.Name, p)
		}
	}

	xlsx.SetActiveSheet(last)
	return xlsx.WriteToBuffer()
}

// Import import excel
func Import(reader io.Reader, sheets map[string]Sheet) error {
	xlsx, err := excelize.OpenReader(reader)
	if nil != err {
		DefaultLogger.Error(err.Error())
		return err
	}
	for k, s := range sheets {
		rows, err := xlsx.Rows(k)
		if nil != err {
			DefaultLogger.Error(err.Error())
			return err
		}
		i := 0
		formatter := make(map[string]Format, 0)

		m := make(map[int]field, 0)
		list := make([]interface{}, 0)
		for rows.Next() {
			row, err := rows.Columns()
			if err != nil {
				DefaultLogger.Error(err)
				return err
			}
			bean := reflect.New(s.T)
			for j, colCell := range row {
				if i == 0 {
					err = handleImportHeader(s.T, &m, &formatter, j, colCell)
					if err != nil {
						DefaultLogger.Error(err)
						return err
					}
				} else {
					if k, ok := m[j]; ok {
						v := bean.Elem()
						if f, ok := formatter[k.Tag]; ok {
							colCell = f.format(colCell)
						}
						convert(colCell, v.Field(k.Idx))
					}
				}
			}
			//fmt.Println(bean.Interface())
			if i != 0 {
				list = append(list, bean.Interface())
			}
			i++
		}
		if nil == s.Result {
			s.Result = new([]interface{})
		}
		*(s.Result) = list
	}

	//b, err := json.Marshal(sheets)
	//if nil != err {
	//	core.Logger.Error(err.Error())
	//	return err
	//}
	//fmt.Println(string(b))
	return nil
}

func handleImportHeader(t reflect.Type, m *map[int]field, formatter *map[string]Format, idx int, col string) error {
	switch t.Kind() {
	case reflect.Ptr:
		for k := 0; k < t.Elem().NumField(); k++ {
			tag := t.Elem().Field(k).Tag.Get("excel_column")
			if tag == col {
				(*m)[idx] = field{
					Idx: k,
					Tag: tag,
				}
				form := t.Elem().Field(idx).Tag.Get("excel_formatter")
				if "" != form {
					var f Formatter
					v, err := url.ParseQuery(form)
					if nil != err {
						DefaultLogger.Error(err.Error())
						return err
					}
					err = DecodeQuery(&f, v)
					if nil != err {
						DefaultLogger.Error(err.Error())
						return err
					}
					if "" != f.Enum {
						list := strings.Split(f.Enum, ",")
						if len(list) > 0 {
							enum := enumFormatter{enum: make(map[string]string, 0)}
							for i := 0; i < len(list); i++ {
								enum.enum[list[i][:strings.Index(list[i], ":")]] = list[i][strings.Index(list[i], ":")+1:]
							}
							(*formatter)[tag] = enum
						}
					}
					if "" != f.Time {
						(*formatter)[tag] = timeImportFormatter{timeLayout: f.Time}
					}
				}
				break
			}
		}
	case reflect.Struct:
		for k := 0; k < t.NumField(); k++ {
			tag := t.Field(k).Tag.Get("excel_column")
			if tag == col {
				(*m)[idx] = field{
					Idx: k,
					Tag: tag,
				}
				form := t.Field(idx).Tag.Get("excel_formatter")
				if "" != form {
					var f Formatter
					v, err := url.ParseQuery(form)
					if nil != err {
						DefaultLogger.Error(err.Error())
						return err
					}
					err = DecodeQuery(&f, v)
					if nil != err {
						DefaultLogger.Error(err.Error())
						return err
					}
					if "" != f.Enum {
						list := strings.Split(f.Enum, ",")
						if len(list) > 0 {
							enum := enumFormatter{enum: make(map[string]string, 0)}
							for i := 0; i < len(list); i++ {
								enum.enum[list[i][strings.Index(list[i], ":")+1:]] = list[i][:strings.Index(list[i], ":")]
							}
							(*formatter)[tag] = enum
						}
					}
					if "" != f.Time {
						(*formatter)[tag] = timeImportFormatter{timeLayout: f.Time}
					}
				}
				break
			}
		}
	default:

	}
	return nil
}

func convert(s string, v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr:
		switch v.Type().String() {
		case "*int":
			val, err := strconv.Atoi(s)
			if nil != err {
				DefaultLogger.Error(err.Error())
				break
			}
			p := new(int)
			*p = val
			v.Set(reflect.ValueOf(p))
		case "*int64":
			val, err := strconv.ParseInt(s, 10, 64)
			if nil != err {
				DefaultLogger.Error(err.Error())
				break
			}
			p := new(int64)
			*p = val
			v.Set(reflect.ValueOf(p))
		case "*string":
			p := new(string)
			*p = s
			v.Set(reflect.ValueOf(p))
		case "*float32":
			val, err := strconv.ParseFloat(s, 32)
			if nil != err {
				DefaultLogger.Error(err.Error())
				break
			}
			p := new(float32)
			*p = float32(val)
			v.Set(reflect.ValueOf(p))
		case "*float64":
			val, err := strconv.ParseFloat(s, 64)
			if nil != err {
				DefaultLogger.Error(err.Error())
				break
			}
			p := new(float64)
			*p = val
			v.Set(reflect.ValueOf(p))
		case "*bool":
			b := false
			if "true" == s {
				b = true
			}
			p := new(bool)
			*p = b
			v.Set(reflect.ValueOf(p))
		default:
			DefaultLogger.Error(v.Type().String(), "not support")
		}
	case reflect.String:
		v.SetString(s)
	case reflect.Int:
		val, err := strconv.ParseInt(s, 10, 32)
		if nil != err {
			DefaultLogger.Error(err.Error())
			break
		}
		v.SetInt(val)
	case reflect.Int64:
		val, err := strconv.ParseInt(s, 10, 64)
		if nil != err {
			DefaultLogger.Error(err.Error())
			break
		}
		v.SetInt(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(s, 64)
		if nil != err {
			DefaultLogger.Error(err.Error())
			break
		}
		v.SetFloat(val)
	case reflect.Bool:
		b := false
		if "true" == s {
			b = true
		}
		v.SetBool(b)
	default:
		fmt.Println(v.Kind())
		DefaultLogger.Error(v.Kind(), "not support")
	}
}
