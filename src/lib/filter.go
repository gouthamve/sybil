package sybil

import "regexp"
import "log"

// FILTERS RETURN TRUE ON MATCH SUCCESS
type NoFilter struct{}

func (f NoFilter) Filter(r *Record) bool {
	return false
}

type IntFilter struct {
	Field   string
	FieldId int16
	Op      string
	Value   int

	table *Table
}

type StrFilter struct {
	Field   string
	FieldId int16
	Op      string
	Value   string
	Regex   *regexp.Regexp

	table *Table
}

type SetFilter struct {
	Field   string
	FieldId int16
	Op      string
	Value   string

	table *Table
}

func (filter IntFilter) Filter(r *Record) bool {
	if r.Populated[filter.FieldId] == 0 {
		return false
	}

	field := r.Ints[filter.FieldId]
	switch filter.Op {
	case "gt":
		return int(field) > int(filter.Value)

	case "lt":
		return int(field) < int(filter.Value)

	case "eq":
		return int(field) == int(filter.Value)

	case "neq":
		return int(field) != int(filter.Value)

	default:

	}

	return false
}

var REGEX_CACHE_SIZE = 100000

func (filter StrFilter) Filter(r *Record) bool {
	if r.Populated[filter.FieldId] == 0 {
		return false
	}

	val := r.Strs[filter.FieldId]
	col := r.block.GetColumnInfo(filter.FieldId)
	filterval := int(col.get_val_id(filter.Value))

	ok := false
	ret := false
	invert := false
	switch filter.Op {
	case "nre":
		invert = true
		fallthrough
	case "re":
		cardinality := len(col.StringTable)
		// we can cache results if the cardinality is reasonably low
		ok = true

		if cardinality < REGEX_CACHE_SIZE {
			ret, ok = col.RCache[int(val)]
			if !ok {
				str_val := col.get_string_for_val(int32(val))
				ret = filter.Regex.MatchString(str_val)
			}
		} else {
			str_val := col.get_string_for_val(int32(val))
			ret = filter.Regex.MatchString(str_val)
		}

		if cardinality < REGEX_CACHE_SIZE && !ok {
			col.RCache[int(val)] = ret
		}

		if invert {
			ret = !ret
		}

	case "eq":
		ret = int(val) == filterval

	case "neq":
		ret = int(val) != filterval

	default:

	}

	return ret
}

func (filter SetFilter) Filter(r *Record) bool {

	col := r.block.GetColumnInfo(filter.FieldId)
	ret := false

	ok := r.Populated[filter.FieldId] == SET_VAL
	if !ok {
		return false
	}

	sets := r.SetMap[filter.FieldId]
	val_id := col.get_val_id(filter.Value)

	switch filter.Op {
	// Check if tag exists
	case "in":
		// Check if tag does not exist
		for _, tag := range sets {
			if tag == val_id {
				return true
			}
		}
	case "nin":
		ret = true
		for _, tag := range sets {
			if tag == val_id {
				return false
			}
		}

	}
	return ret
}

func (t *Table) IntFilter(name string, op string, value int) IntFilter {
	intFilter := IntFilter{Field: name, FieldId: t.get_key_id(name), Op: op, Value: value}
	intFilter.table = t

	return intFilter

}

func (t *Table) StrFilter(name string, op string, value string) StrFilter {
	strFilter := StrFilter{Field: name, FieldId: t.get_key_id(name), Op: op, Value: value}
	strFilter.table = t

	var err error
	if op == "re" || op == "nre" {
		strFilter.Regex, err = regexp.Compile(value)
		if err != nil {
			log.Println("REGEX ERROR", err, "WITH", value)
		}
	}

	return strFilter

}

func (t *Table) SetFilter(name string, op string, value string) SetFilter {
	setFilter := SetFilter{Field: name, FieldId: t.get_key_id(name), Op: op, Value: value}
	setFilter.table = t

	return setFilter

}
