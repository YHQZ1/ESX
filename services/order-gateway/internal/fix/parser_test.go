package fix_test

import (
	"testing"

	"github.com/YHQZ1/esx/services/order-gateway/internal/fix"
)

func TestParseValidNewOrder(t *testing.T) {
	raw := "8=FIX.4.2|9=100|35=D|49=CLIENT|56=ESX|11=ORD001|55=RELIANCE|54=1|38=10|40=2|44=50000|10=000|"

	msg, err := fix.Parse(raw)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if msg.MsgType() != fix.MsgTypeNewOrder {
		t.Fatalf("expected MsgType D, got %s", msg.MsgType())
	}
	if msg.GetString(fix.TagSymbol) != "RELIANCE" {
		t.Fatalf("expected symbol RELIANCE, got %s", msg.GetString(fix.TagSymbol))
	}
	if msg.GetString(fix.TagSide) != fix.SideBuy {
		t.Fatalf("expected side 1, got %s", msg.GetString(fix.TagSide))
	}
	qty, err := msg.GetInt(fix.TagOrderQty)
	if err != nil {
		t.Fatalf("expected quantity, got error: %v", err)
	}
	if qty != 10 {
		t.Fatalf("expected quantity 10, got %d", qty)
	}
	price, err := msg.GetInt(fix.TagPrice)
	if err != nil {
		t.Fatalf("expected price, got error: %v", err)
	}
	if price != 50000 {
		t.Fatalf("expected price 50000, got %d", price)
	}
}

func TestParseMissingBeginString(t *testing.T) {
	_, err := fix.Parse("35=D|55=RELIANCE|54=1|38=10|40=2|44=50000|")
	if err == nil {
		t.Fatal("expected error for missing BeginString")
	}
}

func TestParseMissingMsgType(t *testing.T) {
	_, err := fix.Parse("8=FIX.4.2|9=100|49=CLIENT|55=RELIANCE|")
	if err == nil {
		t.Fatal("expected error for missing MsgType")
	}
}

func TestParseWithSOHDelimiter(t *testing.T) {
	raw := "8=FIX.4.2\x019=100\x0135=D\x0149=CLIENT\x0111=ORD001\x0155=RELIANCE\x0154=2\x0138=5\x0140=2\x0144=50000\x0110=000\x01"
	msg, err := fix.Parse(raw)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if msg.GetString(fix.TagSymbol) != "RELIANCE" {
		t.Fatalf("expected RELIANCE, got %s", msg.GetString(fix.TagSymbol))
	}
	if msg.GetString(fix.TagSide) != fix.SideSell {
		t.Fatalf("expected side 2, got %s", msg.GetString(fix.TagSide))
	}
}

func TestBuildExecutionReport(t *testing.T) {
	report := fix.NewOrderToFIX("order-123", "RELIANCE", fix.SideBuy, "filled", "ORD001")
	if len(report) == 0 {
		t.Fatal("expected non-empty execution report")
	}
	msg, err := fix.Parse(report)
	if err != nil {
		t.Fatalf("execution report should be parseable: %v", err)
	}
	if msg.MsgType() != fix.MsgTypeExecutionReport {
		t.Fatalf("expected MsgType 8, got %s", msg.MsgType())
	}
	if msg.GetString(fix.TagClOrdID) != "ORD001" {
		t.Fatalf("expected ClOrdID ORD001, got %s", msg.GetString(fix.TagClOrdID))
	}
}

func TestBuildRejectReport(t *testing.T) {
	report := fix.RejectToFIX("ORD001", "insufficient funds")
	if len(report) == 0 {
		t.Fatal("expected non-empty reject report")
	}
	msg, err := fix.Parse(report)
	if err != nil {
		t.Fatalf("reject report should be parseable: %v", err)
	}
	if msg.MsgType() != fix.MsgTypeReject {
		t.Fatalf("expected MsgType 3, got %s", msg.MsgType())
	}
}

func TestGetIntMissingTag(t *testing.T) {
	msg := fix.NewMessage()
	msg.Set(fix.TagBeginString, "FIX.4.2")
	_, err := msg.GetInt(fix.TagOrderQty)
	if err == nil {
		t.Fatal("expected error for missing tag")
	}
}

func TestSideConstants(t *testing.T) {
	if fix.SideBuy != "1" {
		t.Fatalf("expected SideBuy=1, got %s", fix.SideBuy)
	}
	if fix.SideSell != "2" {
		t.Fatalf("expected SideSell=2, got %s", fix.SideSell)
	}
}
