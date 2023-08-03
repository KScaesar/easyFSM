package easyFSM

import (
	"errors"
	"fmt"
	"sort"
	"testing"
)

type OrderEventTopic string

const (
	OrderEventTopicPlaced          OrderEventTopic = "Order.Placed"
	OrderEventTopicShipped         OrderEventTopic = "Order.Shipped"
	OrderEventTopicCancelled       OrderEventTopic = "Order.Cancelled"
	OrderEventTopicDelivered       OrderEventTopic = "Order.Delivered"
	OrderEventTopicReturnRequested OrderEventTopic = "Order.ReturnRequested"
	OrderEventTopicCargoReturned   OrderEventTopic = "Order.CargoReturned"
	OrderEventTopicRefundRequested OrderEventTopic = "Order.RefundRequested"
	OrderEventTopicRefunded        OrderEventTopic = "Order.Refunded"
)

type OrderState string

const (
	OrderStateAwaitingPayment  OrderState = "AwaitingPayment"  // 訂單已建立，但尚未收到付款
	OrderStateConfirmed        OrderState = "Confirmed"        // 訂單已經確認，支付和庫存等相關事宜已完成，等待商品出貨
	OrderStateShipped          OrderState = "Shipped"          // 商品已經發貨，正在運送途中
	OrderStateDelivered        OrderState = "Delivered"        // 商品已經成功送達到顧客手中，交易完成
	OrderStateCancelled        OrderState = "Cancelled"        // 訂單在處理過程中被取消，交易不會繼續進行
	OrderStateReturnInProgress OrderState = "ReturnInProgress" // 顧客申請退貨，退貨正在處理中
	OrderStateReturned         OrderState = "Returned"         // 退貨流程已完成，商品已經退回並接收
	OrderStateRefundInProgress OrderState = "RefundInProgress" // 退款正在處理中，將退還付款給顧客
	OrderStateRefunded         OrderState = "Refunded"         // 退款已經完成，付款已退還給顧客
	// OrderStateError            OrderState = "Error"            // 訂單面臨付款錯誤、庫存問題或其他技術問題
)

var OrderStateFSM = NewFSM[OrderEventTopic, OrderState](OrderStateAwaitingPayment).
	DefineTransition(OrderEventTopicPlaced, OrderStateAwaitingPayment, OrderStateConfirmed).
	DefineTransition(OrderEventTopicShipped, OrderStateConfirmed, OrderStateShipped).
	DefineTransition(OrderEventTopicDelivered, OrderStateShipped, OrderStateDelivered).
	DefineTransition(OrderEventTopicCancelled, OrderStateConfirmed, OrderStateCancelled).
	DefineTransition(OrderEventTopicReturnRequested, OrderStateShipped, OrderStateReturnInProgress).
	DefineTransition(OrderEventTopicCargoReturned, OrderStateReturnInProgress, OrderStateReturned).
	DefineTransition(OrderEventTopicRefundRequested, OrderStateReturnInProgress, OrderStateRefundInProgress).
	DefineTransition(OrderEventTopicRefunded, OrderStateRefundInProgress, OrderStateRefunded).
	DefineTransition(OrderEventTopicRefunded, OrderStateReturned, OrderStateRefundInProgress).
	DefineTransition(OrderEventTopicReturnRequested, OrderStateDelivered, OrderStateReturnInProgress)

func TestFSM_StateAll(t *testing.T) {
	expected := []OrderState{OrderStateAwaitingPayment, OrderStateConfirmed, OrderStateShipped, OrderStateDelivered, OrderStateCancelled, OrderStateReturnInProgress, OrderStateReturned, OrderStateRefundInProgress, OrderStateRefunded}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i] < expected[j]
	})

	actual := OrderStateFSM.StateAll()
	sort.Slice(actual, func(i, j int) bool {
		return actual[i] < actual[j]
	})

	fmt.Println(expected)
	fmt.Println(actual)
}

func TestFSM_OnAction(t *testing.T) {
	fsm1 := OrderStateFSM

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

	fsm2 := fsm1.SetCurrent(OrderStateShipped)
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

	fsm3 := fsm1.SetCurrent(OrderStateShipped)
	err = fsm3.OnAction(OrderEventTopic("CloudNetwork.Created"), func(nextState OrderState) error {
		expected := OrderStateReturnInProgress
		if expected != nextState {
			return fmt.Errorf("expected = {%v}, but actual = {%v}", expected, nextState)
		}
		return nil
	})
	if !errors.Is(err, ErrEventNotExist) {
		t.Errorf("NotExistEvent.OnAction: expected = {%v}, but actual = {%v}", ErrEventNotExist, err)
	}
}
