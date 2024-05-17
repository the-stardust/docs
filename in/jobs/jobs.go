package jobs

import (
	"fmt"
	"log"
	"reflect"
	"time"
)

var (
	jobs = make([]*JobDesc, 0)
)

type JobDesc struct {
	Name        string
	WorkNumber  int
	Instance    Job
	Interval    int64 // second
	NextExecute time.Time
}

type Job interface {
	GetJobs() ([]interface{}, error)
	Do(objectID interface{}) error
}

func RegisterWorkConnectJob(instance Job, interval int64, workNum ...int) {
	if interval < 1 {
		interval = 10
	}
	workNumber := 10
	if len(workNum) > 0 {
		workNumber = workNum[0]
	}
	jobs = append(jobs, &JobDesc{Name: reflect.TypeOf(instance).Elem().String(), WorkNumber: workNumber, Instance: instance, Interval: interval, NextExecute: time.Now().Add(time.Duration(interval) * time.Second).Local()})
}

type JobExecute struct {
	ObjectID interface{}
	Func     func(objectID interface{}) error
}

func Run() {
	var jobsChan = make(map[string]chan JobExecute)
	for index := range jobs {
		name := jobs[index].Name
		workNumber := jobs[index].WorkNumber
		jobsChan[name] = make(chan JobExecute, 1000)
		for i := 0; i < workNumber; i++ {
			go func() {
				jobName := name
				for {
					j := <-jobsChan[jobName]
					if j.Func != nil {
						err := j.Func(j.ObjectID)
						if err != nil {
							log.Println(fmt.Sprintf("job-%+v-Do异常,ObjectID:%+vErr:%+v", jobName, j.ObjectID, err.Error()))
						}
					}
				}
			}()
		}
	}

	for {
		now := time.Now().Local()
		for index := range jobs {
			job := jobs[index]
			if job.NextExecute.In(time.Local).Before(now) {
				go func() {
					ids, err := job.Instance.GetJobs()
					if err == nil {
						for i := 0; i < len(ids); i++ {
							objectId := ids[i]
							go func() {
								jobsChan[job.Name] <- JobExecute{Func: job.Instance.Do, ObjectID: objectId}
							}()
						}
					} else {
						log.Println(fmt.Sprintf("job-%+v-GetJobs异常,Job:%+vErr:%+v", job.Name, job, err.Error()))
					}
				}()
				job.NextExecute = now.Add(time.Duration(job.Interval) * time.Second)
			}
		}
		time.Sleep(time.Second)
	}
}
