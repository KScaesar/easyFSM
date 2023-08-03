package easyFSM

import (
	"context"
	"fmt"
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

// The FSM should be placed in the global scope.
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

func ExampleFSM_OnAction() {
	repo := MemoryOrderRepository{}
	ctx := context.Background()

	// UseCaseSuccess:
	// ReturnRequest success!!
	fmt.Printf("UseCaseSuccess:\n")
	err := OrderUseCaseSuccess(repo, ctx)
	if err != nil {
		fmt.Println(err)
	}

	// UseCaseFail:
	// key = {event: Order.ReturnRequested, requiredState: Delivered},
	// but currentState = Cancelled: state not match
	fmt.Printf("\nUseCaseFail:\n")
	err = OrderUseCaseFail(repo, ctx)
	if err != nil {
		fmt.Println(err)
	}
}

type Order struct {
	Id string
	// ... other field
	State OrderState
}

func (o *Order) ReturnRequest() error {
	fsm := OrderStateFSM.CopyFSM(o.State) // copy by vlaue

	return fsm.OnAction(OrderEventTopicReturnRequested, func(nextState OrderState) error {
		o.State = nextState
		fmt.Println("ReturnRequest success!!")
		return nil
	})
}

func OrderUseCaseSuccess(repo OrderRepository, ctx context.Context) error {
	order, err := repo.LockOrderById(ctx, "action_success")
	if err != nil {
		return fmt.Errorf("get obj from db: %w", err)
	}

	return order.ReturnRequest()
}

func OrderUseCaseFail(repo OrderRepository, ctx context.Context) error {
	order, err := repo.LockOrderById(ctx, "action_fail")
	if err != nil {
		return fmt.Errorf("get obj from db: %w", err)
	}

	return order.ReturnRequest()
}

type OrderRepository interface {
	LockOrderById(ctx context.Context, oId string) (Order, error)
}

type MemoryOrderRepository struct{}

func (MemoryOrderRepository) LockOrderById(_ context.Context, oId string) (Order, error) {
	store := map[string]Order{
		"action_success": {
			Id:    "action_success",
			State: OrderStateDelivered,
		},
		"action_fail": {
			Id:    "action_fail",
			State: OrderStateCancelled,
		},
	}
	return store[oId], nil
}
