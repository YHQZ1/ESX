package fix

import (
	"fmt"
	"strconv"
	"strings"
)

const delimiter = "\x01"

type Message struct {
	fields map[int]string
}

func NewMessage() *Message {
	return &Message{fields: make(map[int]string)}
}

func (m *Message) Set(tag int, value string) {
	m.fields[tag] = value
}

func (m *Message) Get(tag int) (string, bool) {
	v, ok := m.fields[tag]
	return v, ok
}

func (m *Message) GetString(tag int) string {
	return m.fields[tag]
}

func (m *Message) GetInt(tag int) (int64, error) {
	v, ok := m.fields[tag]
	if !ok {
		return 0, fmt.Errorf("tag %d not found", tag)
	}
	return strconv.ParseInt(v, 10, 64)
}

func (m *Message) MsgType() string {
	return m.fields[TagMsgType]
}

func Parse(raw string) (*Message, error) {
	raw = strings.ReplaceAll(raw, "|", delimiter)
	parts := strings.Split(raw, delimiter)

	msg := NewMessage()
	for _, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(part, "=")
		if idx < 0 {
			continue
		}
		tagStr := part[:idx]
		value := part[idx+1:]
		tag, err := strconv.Atoi(tagStr)
		if err != nil {
			continue
		}
		msg.Set(tag, value)
	}

	if msg.GetString(TagBeginString) == "" {
		return nil, fmt.Errorf("missing BeginString (tag 8)")
	}
	if msg.GetString(TagMsgType) == "" {
		return nil, fmt.Errorf("missing MsgType (tag 35)")
	}

	return msg, nil
}

func Build(fields map[int]string) string {
	order := []int{
		TagBeginString, TagBodyLength, TagMsgType,
		TagSenderCompID, TagTargetCompID, TagMsgSeqNum, TagSendingTime,
	}

	seen := make(map[int]bool)
	var body strings.Builder

	for _, tag := range order {
		if v, ok := fields[tag]; ok {
			body.WriteString(fmt.Sprintf("%d=%s%s", tag, v, delimiter))
			seen[tag] = true
		}
	}

	for tag, v := range fields {
		if !seen[tag] && tag != TagCheckSum && tag != TagBodyLength {
			body.WriteString(fmt.Sprintf("%d=%s%s", tag, v, delimiter))
		}
	}

	bodyStr := body.String()
	checksum := 0
	for _, c := range bodyStr {
		checksum += int(c)
	}
	checksum = checksum % 256

	return bodyStr + fmt.Sprintf("%d=%03d%s", TagCheckSum, checksum, delimiter)
}

func NewOrderToFIX(orderID, symbol, side, status, clOrdID string) string {
	execType := ExecTypeNew
	ordStatus := OrdStatusNew
	switch status {
	case "filled":
		execType = ExecTypeFilled
		ordStatus = OrdStatusFilled
	case "partial":
		execType = ExecTypePartialFill
		ordStatus = OrdStatusPartialFill
	case "cancelled":
		execType = ExecTypeCancelled
		ordStatus = OrdStatusCancelled
	}

	return Build(map[int]string{
		TagBeginString:  "FIX.4.2",
		TagMsgType:      MsgTypeExecutionReport,
		TagSenderCompID: "ESX",
		TagTargetCompID: "CLIENT",
		TagClOrdID:      clOrdID,
		55:              symbol,
		37:              orderID,
		17:              orderID,
		150:             execType,
		39:              ordStatus,
		TagSide:         side,
	})
}

func RejectToFIX(clOrdID, reason string) string {
	return Build(map[int]string{
		TagBeginString:  "FIX.4.2",
		TagMsgType:      MsgTypeReject,
		TagSenderCompID: "ESX",
		TagTargetCompID: "CLIENT",
		TagClOrdID:      clOrdID,
		58:              reason,
		39:              OrdStatusRejected,
	})
}
