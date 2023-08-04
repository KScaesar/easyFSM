package easyFSM_test

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/KScaesar/easyFSM"
)

func TestFSM_StateAll(t *testing.T) {
	expected := []OrderState{OrderStateAwaitingPayment, OrderStateConfirmed, OrderStateShipped, OrderStateDelivered, OrderStateCancelled, OrderStateReturnInProgress, OrderStateReturned, OrderStateRefundInProgress, OrderStateRefunded}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i] < expected[j]
	})

	actual := OrderStateFSM.StateAll()
	sort.Slice(actual, func(i, j int) bool {
		return actual[i] < actual[j]
	})

	expected_ := fmt.Sprintf("%v", expected)
	actual_ := fmt.Sprintf("%v", actual)
	if expected_ != actual_ {
		t.Errorf("expected = %v, but actual = %v", expected_, actual_)
	}
}

func TestFSM_OnAction(t *testing.T) {
	fsm1 := *OrderStateFSM // get by value

	err := fsm1.OnAction(OrderEventTopicPlaced, func(nextState OrderState) error {
		expected := OrderStateConfirmed
		if expected != nextState {
			return fmt.Errorf("expected = %v, but actual = %v", expected, nextState)
		}
		return nil
	})
	if err != nil {
		t.Errorf("Payed.OnAction: %v", err)
	}

	fsm2 := fsm1.CopyFSM(OrderStateShipped)
	err = fsm2.OnAction(OrderEventTopicReturnRequested, func(nextState OrderState) error {
		expected := OrderStateReturnInProgress
		if expected != nextState {
			return fmt.Errorf("expected = %v, but actual = %v", expected, nextState)
		}
		return nil
	})
	if err != nil {
		t.Errorf("ReturnInProgress.OnAction: %v", err)
	}

	fsm3 := fsm1.CopyFSM(OrderStateShipped)
	err = fsm3.OnAction(OrderEventTopic("CloudNetwork.Created"), func(nextState OrderState) error {
		expected := OrderStateReturnInProgress
		if expected != nextState {
			return fmt.Errorf("expected = {%v}, but actual = {%v}", expected, nextState)
		}
		return nil
	})
	if !errors.Is(err, easyFSM.ErrEventNotExist) {
		t.Errorf("NotExistEvent.OnAction: expected = {%v}, but actual = {%v}", easyFSM.ErrEventNotExist, err)
	}
}
