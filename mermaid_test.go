package easyFSM

import (
	"testing"
)

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
	actual := MermaidGraphByTopDown(OrderStateFSM, nil)

	if expected != actual {
		t.Errorf("expected = %v, but actual = %v", expected, actual)
	}
}
