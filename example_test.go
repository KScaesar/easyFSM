package easyFSM_test

import (
	"context"
	"fmt"

	"github.com/KScaesar/easyFSM"
)

type OrderEvent string

const (
	OrderEventPlaced          OrderEvent = "Order.Placed"
	OrderEventShipped         OrderEvent = "Order.Shipped"
	OrderEventCancelled       OrderEvent = "Order.Cancelled"
	OrderEventDelivered       OrderEvent = "Order.Delivered"
	OrderEventReturnRequested OrderEvent = "Order.ReturnRequested"
	OrderEventCargoReturned   OrderEvent = "Order.CargoReturned"
	OrderEventRefundRequested OrderEvent = "Order.RefundRequested"
	OrderEventRefunded        OrderEvent = "Order.Refunded"
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
	OrderStateError            OrderState = "Error"            // 訂單面臨付款錯誤、庫存問題或其他技術問題
)

// The FSM should be placed in the global scope.
var OrderFSM = easyFSM.NewFSM[OrderEvent, OrderState](OrderStateAwaitingPayment).
	DefineTransition(OrderEventPlaced, OrderStateAwaitingPayment, OrderStateConfirmed).
	DefineTransition(OrderEventShipped, OrderStateConfirmed, OrderStateShipped).
	DefineTransition(OrderEventDelivered, OrderStateShipped, OrderStateDelivered).
	DefineTransition(OrderEventCancelled, OrderStateConfirmed, OrderStateCancelled).
	DefineTransition(OrderEventReturnRequested, OrderStateShipped, OrderStateReturnInProgress).
	DefineTransition(OrderEventCargoReturned, OrderStateReturnInProgress, OrderStateReturned).
	DefineTransition(OrderEventRefundRequested, OrderStateReturnInProgress, OrderStateRefundInProgress).
	DefineTransition(OrderEventRefunded, OrderStateRefundInProgress, OrderStateRefunded).
	DefineTransition(OrderEventRefunded, OrderStateReturned, OrderStateRefundInProgress).
	DefineTransition(OrderEventReturnRequested, OrderStateDelivered, OrderStateReturnInProgress)

func ExampleFSM_OnAction() {
	repo := MemoryOrderRepository{}
	ctx := context.Background()

	fmt.Printf("UseCaseSuccess:\n")
	err := OrderUseCaseSuccess(repo, ctx)
	if err != nil {
		fmt.Println(err)
	}

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
	fsm := OrderFSM.CopyFSM(o.State) // copy by value

	return fsm.OnAction(OrderEventReturnRequested, func(nextState OrderState) error {
		o.State = nextState
		fmt.Println(o.State)
		return nil
	})
}

func OrderUseCaseSuccess(repo OrderRepository, ctx context.Context) error {
	order, err := repo.LockOrderById(ctx, "order_state_is_Delivered")
	if err != nil {
		return fmt.Errorf("get obj from db: %w", err)
	}

	// ReturnInProgress
	return order.ReturnRequest()
}

func OrderUseCaseFail(repo OrderRepository, ctx context.Context) error {
	order, err := repo.LockOrderById(ctx, "order_state_is_Confirmed")
	if err != nil {
		return fmt.Errorf("get obj from db: %w", err)
	}

	// key = {event: Order.ReturnRequested, requiredState: Delivered}, but currentState = Confirmed: state not match
	return order.ReturnRequest()
}

type OrderRepository interface {
	LockOrderById(ctx context.Context, oId string) (Order, error)
}

type MemoryOrderRepository struct{}

func (MemoryOrderRepository) LockOrderById(_ context.Context, oId string) (Order, error) {
	store := map[string]Order{
		"order_state_is_Delivered": {
			Id:    "order_state_is_Delivered",
			State: OrderStateDelivered,
		},
		"order_state_is_Confirmed": {
			Id:    "order_state_is_Confirmed",
			State: OrderStateConfirmed,
		},
	}
	return store[oId], nil
}
