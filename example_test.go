package easyFSM

import (
	"context"
	"fmt"
)

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
	// key = {event: Order.ReturnRequested, requiredState: Delivered}, but currentState = Cancelled: state not match
	fmt.Printf("\nUseCaseFail:\n")
	err = OrderUseCaseFail(repo, ctx)
	if err != nil {
		fmt.Println(err)
	}
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
		return err
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

type Order struct {
	Id string
	// ... other field
	State OrderState
}

func (o *Order) ReturnRequest() error {
	fsm := OrderStateFSM.SetCurrent(o.State)
	return fsm.OnAction(OrderEventTopicReturnRequested, func(nextState OrderState) error {
		o.State = nextState
		fmt.Println("ReturnRequest success!!")
		return nil
	})
}
