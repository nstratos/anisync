package anisync

import (
	"github.com/nstratos/go-myanimelist/mal"
)

type Status int

const (
	Unknown Status = iota
	Current
	Planned
	Completed
	OnHold
	Dropped
)

// Possible Anime status values.
//
// 	currently-watching    <->    1
// 	plan-to-watch         <->    6
// 	completed             <->    2
// 	on-hold               <->    3
// 	dropped               <->    4
const (
	StatusCurrentlyWatching = "currently-watching"
	StatusPlanToWatch       = "plan-to-watch"
	StatusCompleted         = "completed"
	StatusOnHold            = "on-hold"
	StatusDropped           = "dropped"
)

func fromHBStatus(status string) Status {
	switch status {
	case StatusCurrentlyWatching:
		return Current
	case StatusPlanToWatch:
		return Planned
	case StatusCompleted:
		return Completed
	case StatusOnHold:
		return OnHold
	case StatusDropped:
		return Dropped
	default:
		return Unknown
	}
}

func FromMALStatus(status mal.Status) Status {
	switch status {
	case mal.Current:
		return Current
	case mal.Planned:
		return Planned
	case mal.Completed:
		return Completed
	case mal.OnHold:
		return OnHold
	case mal.Dropped:
		return Dropped
	default:
		return Unknown
	}
}

func toMALStatus(status Status) mal.Status {
	switch status {
	case Current:
		return mal.Current
	case Planned:
		return mal.Planned
	case Completed:
		return mal.Completed
	case OnHold:
		return mal.OnHold
	case Dropped:
		return mal.Dropped
	default:
		return 0
	}
}

//1/watching, 2/completed, 3/onhold, 4/dropped, 6/plantowatch
//type Status interface {
//	Key() string
//	Name() string
//	Value() int
//}
//
//func NewStatus(v int) (Status, error) {
//	var s Status
//	switch v {
//	case 1:
//		return s.(CurrentlyWatching), nil
//	case 2:
//		return s.(Completed), nil
//	case 3:
//		return s.(OnHold), nil
//	case 4:
//		return s.(Dropped), nil
//	case 6:
//		return s.(PlanToWatch), nil
//	default:
//		return nil, fmt.Errorf("no status value provided")
//	}
//}
//
//type CurrentlyWatching string
//
//func (s CurrentlyWatching) Name() string   { return "Currently Watching" }
//func (s CurrentlyWatching) Key() string    { return "currently-watching" }
//func (s CurrentlyWatching) Value() int     { return 1 }
//func (s CurrentlyWatching) String() string { return s.Key() }
//
//type PlanToWatch string
//
//func (s PlanToWatch) Name() string   { return "Plan to watch" }
//func (s PlanToWatch) Key() string    { return "plan-to-watch" }
//func (s PlanToWatch) Value() int     { return 6 }
//func (s PlanToWatch) String() string { return s.Key() }
//
//type Completed string
//
//func (s Completed) Name() string   { return "Completed" }
//func (s Completed) Key() string    { return "completed" }
//func (s Completed) Value() int     { return 2 }
//func (s Completed) String() string { return s.Key() }
//
//type OnHold string
//
//func (s OnHold) Name() string   { return "On hold" }
//func (s OnHold) Key() string    { return "on-hold" }
//func (s OnHold) Value() int     { return 3 }
//func (s OnHold) String() string { return s.Key() }
//
//type Dropped string
//
//func (s Dropped) Name() string   { return "Dropped" }
//func (s Dropped) Key() string    { return "dropped" }
//func (s Dropped) Value() int     { return 4 }
//func (s Dropped) String() string { return s.Key() }
