//go:build co

//go:generate go install github.com/goghcrow/go-co/cmd/cogen@main
//go:generate cogen

package microthread

import (
	"math/rand"
	"time"

	. "github.com/goghcrow/go-co"
)

type Soldier struct {
	AnimateReload *Signal
}

func NewSoldier(animateReload *Signal) *Soldier {
	return &Soldier{
		AnimateReload: animateReload,
	}
}

func (s *Soldier) Init(shed *Sched) {
	shed.AddTask(s.Patrol())
}

func (s *Soldier) Patrol() (_ Iter[State]) {
	for s.Alive() {
		if s.CanSeeTarget() {
			YieldFrom(s.Attack())
		} else if s.InReloadStation() {
			Yield(s.AnimateReload)
		} else {
			s.MoveTowardsNextWayPoint()
			Yield(WaitFor(time.Second))
		}
	}
	return
}

func (s *Soldier) Attack() (_ Iter[State]) {
	for s.TargetAlive() && s.CanSeeTarget() {
		s.AimAtTarget()
		s.Fire()
		Yield(WaitFor(time.Second))
	}
	return
}

func (s *Soldier) InReloadStation() bool {
	inReloadStation := rand.Intn(5) == 0
	if inReloadStation {
		println("InReloadStation")
	}
	return inReloadStation
}

func (s *Soldier) Alive() bool {
	return true
}

func (s *Soldier) CanSeeTarget() bool {
	return rand.Intn(3) != 0
}

func (s *Soldier) TargetAlive() bool {
	return rand.Intn(2) == 0
}

func (s *Soldier) AimAtTarget() {
	println("AimAtTarget")
}

func (s *Soldier) Fire() {
	println("Fire")
}

func (s *Soldier) MoveTowardsNextWayPoint() {
	println("MoveTowardsNextWayPoint")
}
