package jobs

import (
	"reflect"
)

var (
	onceJobs = make([]*OnceJobDesc, 0)
)

type OnceJobDesc struct {
	Name     string
	Instance OnceJob
}
type OnceJob interface {
	Do()
}

func RegisterOnceJob(instance OnceJob) {
	onceJobs = append(onceJobs, &OnceJobDesc{Name: reflect.TypeOf(instance).Elem().String(), Instance: instance})
}
func OnceRun() {
	for _, v := range onceJobs {
		go v.Instance.Do()
	}
}
