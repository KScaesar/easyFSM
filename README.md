# Tutorial

I really like the concept of Domain-Driven Design (DDD), but I often struggle with handling state changes in domain objects.  

I created this package with the motivation to simplify state management and use Mermaid for visualizing state transitions.  

The usual way of managing state changes in domain objects often means dealing with a bunch of if conditions, and that can be error-prone.  

This package aims to provide a more straightforward and organized way to manage states, making it easier for developers to understand and maintain their code.  

By visualizing state transitions using Mermaid, users can gain a clearer understanding of the system's behavior and enhance their overall development experience.  

- [Tutorial](#tutorial)
	- [Install](#install)
	- [Example](#example)
		- [Step 1: Define state-diagram on the Mermaid website](#step-1-define-state-diagram-on-the-mermaid-website)
		- [Step 2: Define transitions in Golang code](#step-2-define-transitions-in-golang-code)
		- [Step 3: Verify that Golang FSM meets expectations](#step-3-verify-that-golang-fsm-meets-expectations)
		- [Step 4: Call the domain object method in Domain-Driven Design (DDD)](#step-4-call-the-domain-object-method-in-domain-driven-design-ddd)

## Install

```shell
go get github.com/KScaesar/easyFSM
```

## Example

### Step 1: Define state-diagram on the Mermaid website

- [state-diagram-docs](https://github.com/mermaid-js/mermaid#state-diagram-docs---live-editor)  
- [live editor](https://mermaid.live/edit)  

```mermaid
graph TD
    AwaitingPayment --> |Order.Placed| Confirmed
    Confirmed --> |Order.Shipped| Shipped
    Shipped --> |Order.Delivered| Delivered
    Confirmed --> |Order.Cancelled| Cancelled
    Shipped --> |Order.ReturnRequested| ReturnInProgress
    ReturnInProgress --> |Order.CargoReturned| Returned
    ReturnInProgress --> |Order.RefundRequested| RefundInProgress
    RefundInProgress --> |Order.Refunded| Refunded
    Returned --> |Order.Refunded| RefundInProgress
    Delivered --> |Order.ReturnRequested| ReturnInProgress
```

### Step 2: Define transitions in Golang code

1. The FSM should be placed in the global scope.
2. When importing the package, transitions should be added using the DefineTransition function during the initialization step.
3. The event triggering the transition.

```go
// Parameters:
// - event: The event triggering the transition.
// - src: The source state from which the transition is allowed.
// - dest: The destination state to which the FSM will transition when the event occurs in the source state.
func (fsm FSM[E, S]) DefineTransition(event E, src, dest S) FSM[E, S]
```

```go
// The FSM should be placed in the global scope.
var OrderFSM = NewFSM[OrderEvent, OrderState](OrderStateAwaitingPayment).
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
	// OrderStateError            OrderState = "Error"            // 訂單面臨付款錯誤、庫存問題或其他技術問題
)
```

### Step 3: Verify that Golang FSM meets expectations

```go
func TestMermaidGraphByTopDown(t *testing.T) {
	expected := `
graph TD
  AwaitingPayment --> |Order.Placed| Confirmed
  Confirmed --> |Order.Shipped| Shipped
  Shipped --> |Order.Delivered| Delivered
  Confirmed --> |Order.Cancelled| Cancelled
  Shipped --> |Order.ReturnRequested| ReturnInProgress
  ReturnInProgress --> |Order.CargoReturned| Returned
  ReturnInProgress --> |Order.RefundRequested| RefundInProgress
  RefundInProgress --> |Order.Refunded| Refunded
  Returned --> |Order.Refunded| RefundInProgress
  Delivered --> |Order.ReturnRequested| ReturnInProgress
`
	actual := MermaidGraphByTopDown(OrderFSM, nil)

	if expected != actual {
		t.Errorf("expected = %v, but actual = %v", expected, actual)
	}
}
```

### Step 4: Call the domain object method in Domain-Driven Design (DDD)

[playground](https://go.dev/play/p/WJVVyZNRnTF)

```go
type Order struct {
	Id string
	// ... other field
	State OrderState
}

func (o *Order) ReturnRequest() error {
	fsm := OrderStateFSM.CopyFSM(o.State) // copy by value

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
```
