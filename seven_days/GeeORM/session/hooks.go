package session

import (
	"GeeORM/log"
	"reflect"
)

// hooks constants
const (
	BeforeQuery  = "BeforeQuery"
	AfterQuery   = "AfterQuery"
	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"
	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"
)

// CallMethod 调用注册的钩子
func (s *Session) CallMethod(method string, value interface{}) {
	fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)
	if value != nil {
		fm = reflect.ValueOf(value).MethodByName(method)
	}
	// 入参，每个钩子的入参均为 *Session
	param := []reflect.Value{reflect.ValueOf(s)}
	if fm.IsValid() {
		if v := fm.Call(param); len(v) > 0 {
			if err, ok := v[0].Interface().(error); ok {
				log.Error(err)
			}
		}
	}
}

// // interface 实现
// type IAfterQuery interface {
//     AfterQuery(s *Session) error
// }

// type IBeforeInsert interface {
//     BeforeInsert(s *Session) error
// }

// // CallMethod calls the registered hooks
// func (s *Session) CallMethod(method string, value interface{}) {
// 	param := reflect.ValueOf(value)
//     switch method {
//     case AfterQuery:
//         if i, ok := param.Interface().(IAfterQuery); ok {
//             i.AfterQuery(s)
//         }
//     case BeforeInsert:
//         if i, ok := param.Interface().(IBeforeInsert); ok {
//             i.BeforeInsert(s)
//         }
//     default:
//         panic("unsupported hook method")
//     }
// 	return
// }
