package mclock

import "testing"

// This test checks basic functionality of Alarm.
func TestAlarm(t *testing.T) {
	clk := new(Simulated)
	clk.Run(20)
	a := NewAlarm(clk)

	a.Schedule(clk.Now() + 10)
	if recv(a.C()) {
		t.Fatal("Alarm fired before scheduled deadline")
	}
	if ntimers := clk.ActiveTimers(); ntimers != 1 {
		t.Fatal("clock has", ntimers, "active timers, want", 1)
	}
	clk.Run(5)
	if recv(a.C()) {
		t.Fatal("Alarm fired too early")
	}

	clk.Run(5)
	if !recv(a.C()) {
		t.Fatal("Alarm did not fire")
	}
	if recv(a.C()) {
		t.Fatal("Alarm fired twice")
	}
	if ntimers := clk.ActiveTimers(); ntimers != 0 {
		t.Fatal("clock has", ntimers, "active timers, want", 0)
	}

	a.Schedule(clk.Now() + 5)
	if recv(a.C()) {
		t.Fatal("Alarm fired before scheduled deadline when scheduling the second event")
	}

	clk.Run(5)
	if !recv(a.C()) {
		t.Fatal("Alarm did not fire when scheduling the second event")
	}
	if recv(a.C()) {
		t.Fatal("Alarm fired twice when scheduling the second event")
	}
}

// This test checks that scheduling an Alarm to an earlier time than the
// one already scheduled works properly.
func TestAlarmScheduleEarlier(t *testing.T) {
	clk := new(Simulated)
	clk.Run(20)
	a := NewAlarm(clk)

	a.Schedule(clk.Now() + 50)
	clk.Run(5)
	a.Schedule(clk.Now() + 1)
	clk.Run(3)
	if !recv(a.C()) {
		t.Fatal("Alarm did not fire")
	}
}

// This test checks that scheduling an Alarm to a later time than the
// one already scheduled works properly.
func TestAlarmScheduleLater(t *testing.T) {
	clk := new(Simulated)
	clk.Run(20)
	a := NewAlarm(clk)

	a.Schedule(clk.Now() + 50)
	clk.Run(5)
	a.Schedule(clk.Now() + 100)
	clk.Run(50)
	if !recv(a.C()) {
		t.Fatal("Alarm did not fire")
	}
}

// This test checks that scheduling an Alarm in the past makes it fire immediately.
func TestAlarmNegative(t *testing.T) {
	clk := new(Simulated)
	clk.Run(50)
	a := NewAlarm(clk)

	a.Schedule(-1)
	clk.Run(1) // needed to process timers
	if !recv(a.C()) {
		t.Fatal("Alarm did not fire for negative time")
	}
}

func recv(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
