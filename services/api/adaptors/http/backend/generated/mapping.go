// Code generated from OpenAPI, Protobuf and explicit ACL mappings. DO NOT EDIT.
package generated

import (
	"encoding/json"
	"fmt"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1"
	collection "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/collection"
	communication "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/communication"
	loyalty "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/loyalty"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/menu"
	ordering "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/ordering"
	preparation "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/preparation"
	"google.golang.org/protobuf/proto"
)

func Supported(operation string) bool {
	switch operation {
	case "addLine", "addOffer", "changePrice", "changeQuantity", "collectOrder", "completePreparation", "createDrink", "createEdition", "createOrder", "getAccount", "getDrink", "getEdition", "getNotification", "getOrder", "getPickup", "getReward", "getTicket", "listAccounts", "listDrinks", "listEditions", "listNotifications", "listOrders", "listPickups", "listRewards", "listTickets", "placeOrder", "publishDrink", "publishEdition", "redeemReward", "reviseDrink", "startPreparation":
		return true
	}
	return false
}
func Command(operation string, data []byte, metadata *pb.CommandMetadata) (pb.CommandEnvelope, error) {
	switch operation {
	case "createDrink":
		var input struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &menu.Command{Metadata: metadata, Payload: &menu.Command_CreateDrink{CreateDrink: &menu.CreateDrink{Name: &input.Name}}}, nil
	case "reviseDrink":
		var input struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &menu.Command{Metadata: metadata, Payload: &menu.Command_ReviseDrink{ReviseDrink: &menu.ReviseDrink{Name: &input.Name}}}, nil
	case "publishDrink":
		var input struct {
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &menu.Command{Metadata: metadata, Payload: &menu.Command_PublishDrink{PublishDrink: &menu.PublishDrink{}}}, nil
	case "createEdition":
		var input struct {
			Currency string `json:"currency"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &menu.Command{Metadata: metadata, Payload: &menu.Command_CreateEdition{CreateEdition: &menu.CreateEdition{Currency: &input.Currency}}}, nil
	case "addOffer":
		var input struct {
			Code          string `json:"code"`
			DrinkId       string `json:"drinkId"`
			DrinkRevision uint64 `json:"drinkRevision"`
			Minor         int64  `json:"minor"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &menu.Command{Metadata: metadata, Payload: &menu.Command_AddOffer{AddOffer: &menu.AddOffer{Code: &input.Code, DrinkId: &input.DrinkId, DrinkRevision: &input.DrinkRevision, Minor: &input.Minor}}}, nil
	case "changePrice":
		var input struct {
			Code  string `json:"code"`
			Minor int64  `json:"minor"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &menu.Command{Metadata: metadata, Payload: &menu.Command_ChangePrice{ChangePrice: &menu.ChangePrice{Code: &input.Code, Minor: &input.Minor}}}, nil
	case "publishEdition":
		var input struct {
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &menu.Command{Metadata: metadata, Payload: &menu.Command_PublishEdition{PublishEdition: &menu.PublishEdition{}}}, nil
	case "createOrder":
		var input struct {
			CustomerId string `json:"customerId"`
			EditionId  string `json:"editionId"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &ordering.Command{Metadata: metadata, Payload: &ordering.Command_CreateOrder{CreateOrder: &ordering.CreateOrder{CustomerId: &input.CustomerId, EditionId: &input.EditionId}}}, nil
	case "addLine":
		var input struct {
			LineId    string `json:"lineId"`
			EditionId string `json:"editionId"`
			OfferCode string `json:"offerCode"`
			Quantity  int32  `json:"quantity"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &ordering.Command{Metadata: metadata, Payload: &ordering.Command_AddLine{AddLine: &ordering.AddLine{LineId: &input.LineId, EditionId: &input.EditionId, OfferCode: &input.OfferCode, Quantity: &input.Quantity}}}, nil
	case "changeQuantity":
		var input struct {
			LineId   string `json:"lineId"`
			Quantity int32  `json:"quantity"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &ordering.Command{Metadata: metadata, Payload: &ordering.Command_ChangeQuantity{ChangeQuantity: &ordering.ChangeQuantity{LineId: &input.LineId, Quantity: &input.Quantity}}}, nil
	case "placeOrder":
		var input struct {
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &ordering.Command{Metadata: metadata, Payload: &ordering.Command_PlaceOrder{PlaceOrder: &ordering.PlaceOrder{}}}, nil
	case "startPreparation":
		var input struct {
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &preparation.Command{Metadata: metadata, Payload: &preparation.Command_StartPreparation{StartPreparation: &preparation.StartPreparation{}}}, nil
	case "completePreparation":
		var input struct {
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &preparation.Command{Metadata: metadata, Payload: &preparation.Command_CompletePreparation{CompletePreparation: &preparation.CompletePreparation{}}}, nil
	case "collectOrder":
		var input struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &collection.Command{Metadata: metadata, Payload: &collection.Command_CollectOrder{CollectOrder: &collection.CollectOrder{Code: &input.Code}}}, nil
	case "redeemReward":
		var input struct {
			OrderId string `json:"orderId"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, err
		}
		return &loyalty.Command{Metadata: metadata, Payload: &loyalty.Command_RedeemReward{RedeemReward: &loyalty.RedeemReward{OrderId: &input.OrderId}}}, nil
	}
	return nil, fmt.Errorf("unknown command operation")
}
func Query(operation, id string, page *pb.PageRequest) (proto.Message, error) {
	switch operation {
	case "listDrinks":
		return &menu.Query{Payload: &menu.Query_ListDrinks{ListDrinks: &menu.ListDrinks{Page: page}}}, nil
	case "getDrink":
		return &menu.Query{Payload: &menu.Query_GetDrink{GetDrink: &menu.GetDrink{Id: id}}}, nil
	case "listEditions":
		return &menu.Query{Payload: &menu.Query_ListEditions{ListEditions: &menu.ListEditions{Page: page}}}, nil
	case "getEdition":
		return &menu.Query{Payload: &menu.Query_GetEdition{GetEdition: &menu.GetEdition{Id: id}}}, nil
	case "listOrders":
		return &ordering.Query{Payload: &ordering.Query_ListOrders{ListOrders: &ordering.ListOrders{Page: page}}}, nil
	case "getOrder":
		return &ordering.Query{Payload: &ordering.Query_GetOrder{GetOrder: &ordering.GetOrder{Id: id}}}, nil
	case "listTickets":
		return &preparation.Query{Payload: &preparation.Query_ListTickets{ListTickets: &preparation.ListTickets{Page: page}}}, nil
	case "getTicket":
		return &preparation.Query{Payload: &preparation.Query_GetTicket{GetTicket: &preparation.GetTicket{Id: id}}}, nil
	case "listPickups":
		return &collection.Query{Payload: &collection.Query_ListPickups{ListPickups: &collection.ListPickups{Page: page}}}, nil
	case "getPickup":
		return &collection.Query{Payload: &collection.Query_GetPickup{GetPickup: &collection.GetPickup{Id: id}}}, nil
	case "listAccounts":
		return &loyalty.Query{Payload: &loyalty.Query_ListAccounts{ListAccounts: &loyalty.ListAccounts{Page: page}}}, nil
	case "getAccount":
		return &loyalty.Query{Payload: &loyalty.Query_GetAccount{GetAccount: &loyalty.GetAccount{Id: id}}}, nil
	case "listRewards":
		return &loyalty.Query{Payload: &loyalty.Query_ListRewards{ListRewards: &loyalty.ListRewards{Page: page}}}, nil
	case "getReward":
		return &loyalty.Query{Payload: &loyalty.Query_GetReward{GetReward: &loyalty.GetReward{Id: id}}}, nil
	case "listNotifications":
		return &communication.Query{Payload: &communication.Query_ListNotifications{ListNotifications: &communication.ListNotifications{Page: page}}}, nil
	case "getNotification":
		return &communication.Query{Payload: &communication.Query_GetNotification{GetNotification: &communication.GetNotification{Id: id}}}, nil
	}
	return nil, fmt.Errorf("unknown query operation")
}
func QueryReply(request pb.Request, reply pb.Reply) (any, bool, *string, error) {
	switch request := request.(type) {
	case *menu.Request:
		reply, ok := reply.(*menu.Reply)
		if !ok || request.GetQuery() == nil {
			return nil, false, nil, fmt.Errorf("reply owner disagrees with query")
		}
		switch value := request.GetQuery().Payload.(type) {
		case *menu.Query_ListDrinks:
			result, ok := reply.Payload.(*menu.Reply_Drinks)
			if !ok || result.Drinks == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			items := make([]any, 0, len(result.Drinks.Items))
			for _, item := range result.Drinks.Items {
				object, err := mapmenuLoadedDrink(item)
				if err != nil {
					return nil, false, nil, err
				}
				items = append(items, object)
			}
			paged := value.ListDrinks.Page != nil
			if result.Drinks.Paged != paged {
				return nil, false, nil, fmt.Errorf("page shape mismatch")
			}
			return items, paged, result.Drinks.NextId, nil
		case *menu.Query_GetDrink:
			result, ok := reply.Payload.(*menu.Reply_Drink)
			if !ok || result.Drink == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			item := result.Drink
			if !item.Exists || item.State == nil || item.State.Id != value.GetDrink.Id {
				return nil, false, nil, fmt.Errorf("query identity mismatch")
			}
			object, err := mapmenuLoadedDrink(item)
			return object, false, nil, err
		case *menu.Query_ListEditions:
			result, ok := reply.Payload.(*menu.Reply_Editions)
			if !ok || result.Editions == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			items := make([]any, 0, len(result.Editions.Items))
			for _, item := range result.Editions.Items {
				object, err := mapmenuLoadedEdition(item)
				if err != nil {
					return nil, false, nil, err
				}
				items = append(items, object)
			}
			paged := value.ListEditions.Page != nil
			if result.Editions.Paged != paged {
				return nil, false, nil, fmt.Errorf("page shape mismatch")
			}
			return items, paged, result.Editions.NextId, nil
		case *menu.Query_GetEdition:
			result, ok := reply.Payload.(*menu.Reply_Edition)
			if !ok || result.Edition == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			item := result.Edition
			if !item.Exists || item.State == nil || item.State.Id != value.GetEdition.Id {
				return nil, false, nil, fmt.Errorf("query identity mismatch")
			}
			object, err := mapmenuLoadedEdition(item)
			return object, false, nil, err
		}
	case *ordering.Request:
		reply, ok := reply.(*ordering.Reply)
		if !ok || request.GetQuery() == nil {
			return nil, false, nil, fmt.Errorf("reply owner disagrees with query")
		}
		switch value := request.GetQuery().Payload.(type) {
		case *ordering.Query_ListOrders:
			result, ok := reply.Payload.(*ordering.Reply_Orders)
			if !ok || result.Orders == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			items := make([]any, 0, len(result.Orders.Items))
			for _, item := range result.Orders.Items {
				object, err := maporderingLoadedOrder(item)
				if err != nil {
					return nil, false, nil, err
				}
				items = append(items, object)
			}
			paged := value.ListOrders.Page != nil
			if result.Orders.Paged != paged {
				return nil, false, nil, fmt.Errorf("page shape mismatch")
			}
			return items, paged, result.Orders.NextId, nil
		case *ordering.Query_GetOrder:
			result, ok := reply.Payload.(*ordering.Reply_Order)
			if !ok || result.Order == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			item := result.Order
			if !item.Exists || item.State == nil || item.State.Id != value.GetOrder.Id {
				return nil, false, nil, fmt.Errorf("query identity mismatch")
			}
			object, err := maporderingLoadedOrder(item)
			return object, false, nil, err
		}
	case *preparation.Request:
		reply, ok := reply.(*preparation.Reply)
		if !ok || request.GetQuery() == nil {
			return nil, false, nil, fmt.Errorf("reply owner disagrees with query")
		}
		switch value := request.GetQuery().Payload.(type) {
		case *preparation.Query_ListTickets:
			result, ok := reply.Payload.(*preparation.Reply_Tickets)
			if !ok || result.Tickets == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			items := make([]any, 0, len(result.Tickets.Items))
			for _, item := range result.Tickets.Items {
				object, err := mappreparationLoadedTicket(item)
				if err != nil {
					return nil, false, nil, err
				}
				items = append(items, object)
			}
			paged := value.ListTickets.Page != nil
			if result.Tickets.Paged != paged {
				return nil, false, nil, fmt.Errorf("page shape mismatch")
			}
			return items, paged, result.Tickets.NextId, nil
		case *preparation.Query_GetTicket:
			result, ok := reply.Payload.(*preparation.Reply_Ticket)
			if !ok || result.Ticket == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			item := result.Ticket
			if !item.Exists || item.State == nil || item.State.Id != value.GetTicket.Id {
				return nil, false, nil, fmt.Errorf("query identity mismatch")
			}
			object, err := mappreparationLoadedTicket(item)
			return object, false, nil, err
		}
	case *collection.Request:
		reply, ok := reply.(*collection.Reply)
		if !ok || request.GetQuery() == nil {
			return nil, false, nil, fmt.Errorf("reply owner disagrees with query")
		}
		switch value := request.GetQuery().Payload.(type) {
		case *collection.Query_ListPickups:
			result, ok := reply.Payload.(*collection.Reply_Pickups)
			if !ok || result.Pickups == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			items := make([]any, 0, len(result.Pickups.Items))
			for _, item := range result.Pickups.Items {
				object, err := mapcollectionLoadedPickup(item)
				if err != nil {
					return nil, false, nil, err
				}
				items = append(items, object)
			}
			paged := value.ListPickups.Page != nil
			if result.Pickups.Paged != paged {
				return nil, false, nil, fmt.Errorf("page shape mismatch")
			}
			return items, paged, result.Pickups.NextId, nil
		case *collection.Query_GetPickup:
			result, ok := reply.Payload.(*collection.Reply_Pickup)
			if !ok || result.Pickup == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			item := result.Pickup
			if !item.Exists || item.State == nil || item.State.Id != value.GetPickup.Id {
				return nil, false, nil, fmt.Errorf("query identity mismatch")
			}
			object, err := mapcollectionLoadedPickup(item)
			return object, false, nil, err
		}
	case *loyalty.Request:
		reply, ok := reply.(*loyalty.Reply)
		if !ok || request.GetQuery() == nil {
			return nil, false, nil, fmt.Errorf("reply owner disagrees with query")
		}
		switch value := request.GetQuery().Payload.(type) {
		case *loyalty.Query_ListAccounts:
			result, ok := reply.Payload.(*loyalty.Reply_Accounts)
			if !ok || result.Accounts == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			items := make([]any, 0, len(result.Accounts.Items))
			for _, item := range result.Accounts.Items {
				object, err := maployaltyLoadedAccount(item)
				if err != nil {
					return nil, false, nil, err
				}
				items = append(items, object)
			}
			paged := value.ListAccounts.Page != nil
			if result.Accounts.Paged != paged {
				return nil, false, nil, fmt.Errorf("page shape mismatch")
			}
			return items, paged, result.Accounts.NextId, nil
		case *loyalty.Query_GetAccount:
			result, ok := reply.Payload.(*loyalty.Reply_Account)
			if !ok || result.Account == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			item := result.Account
			if !item.Exists || item.State == nil || item.State.Id != value.GetAccount.Id {
				return nil, false, nil, fmt.Errorf("query identity mismatch")
			}
			object, err := maployaltyLoadedAccount(item)
			return object, false, nil, err
		case *loyalty.Query_ListRewards:
			result, ok := reply.Payload.(*loyalty.Reply_Rewards)
			if !ok || result.Rewards == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			items := make([]any, 0, len(result.Rewards.Items))
			for _, item := range result.Rewards.Items {
				object, err := maployaltyLoadedReward(item)
				if err != nil {
					return nil, false, nil, err
				}
				items = append(items, object)
			}
			paged := value.ListRewards.Page != nil
			if result.Rewards.Paged != paged {
				return nil, false, nil, fmt.Errorf("page shape mismatch")
			}
			return items, paged, result.Rewards.NextId, nil
		case *loyalty.Query_GetReward:
			result, ok := reply.Payload.(*loyalty.Reply_Reward)
			if !ok || result.Reward == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			item := result.Reward
			if !item.Exists || item.State == nil || item.State.Id != value.GetReward.Id {
				return nil, false, nil, fmt.Errorf("query identity mismatch")
			}
			object, err := maployaltyLoadedReward(item)
			return object, false, nil, err
		}
	case *communication.Request:
		reply, ok := reply.(*communication.Reply)
		if !ok || request.GetQuery() == nil {
			return nil, false, nil, fmt.Errorf("reply owner disagrees with query")
		}
		switch value := request.GetQuery().Payload.(type) {
		case *communication.Query_ListNotifications:
			result, ok := reply.Payload.(*communication.Reply_Notifications)
			if !ok || result.Notifications == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			items := make([]any, 0, len(result.Notifications.Items))
			for _, item := range result.Notifications.Items {
				object, err := mapcommunicationLoadedNotification(item)
				if err != nil {
					return nil, false, nil, err
				}
				items = append(items, object)
			}
			paged := value.ListNotifications.Page != nil
			if result.Notifications.Paged != paged {
				return nil, false, nil, fmt.Errorf("page shape mismatch")
			}
			return items, paged, result.Notifications.NextId, nil
		case *communication.Query_GetNotification:
			result, ok := reply.Payload.(*communication.Reply_Notification)
			if !ok || result.Notification == nil {
				return nil, false, nil, fmt.Errorf("reply disagrees with query")
			}
			item := result.Notification
			if !item.Exists || item.State == nil || item.State.Id != value.GetNotification.Id {
				return nil, false, nil, fmt.Errorf("query identity mismatch")
			}
			object, err := mapcommunicationLoadedNotification(item)
			return object, false, nil, err
		}
	}
	return nil, false, nil, fmt.Errorf("unknown query payload")
}
func mapcollectionPickup(value *collection.Pickup) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Pickup DTO")
	}
	result := map[string]any{}
	result["id"] = value.Id
	result["orderId"] = value.OrderId
	result["customerId"] = value.CustomerId
	result["code"] = value.Code
	result["status"] = value.Status
	return result, nil
}
func mapcollectionLoadedPickup(value *collection.LoadedPickup) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing LoadedPickup DTO")
	}
	result := map[string]any{}
	result["exists"] = value.Exists
	result["version"] = value.Version
	itemState, err := mapcollectionPickup(value.State)
	if err != nil {
		return nil, err
	}
	result["state"] = itemState
	if !value.Exists || value.State == nil {
		return nil, fmt.Errorf("invalid loaded DTO")
	}
	return result, nil
}
func mapcollectionPickups(value *collection.Pickups) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Pickups DTO")
	}
	result := map[string]any{}
	itemsItems := make([]any, 0, len(value.Items))
	for _, child := range value.Items {
		item, err := mapcollectionLoadedPickup(child)
		if err != nil {
			return nil, err
		}
		itemsItems = append(itemsItems, item)
	}
	result["items"] = itemsItems
	return result, nil
}
func mapcommunicationNotification(value *communication.Notification) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Notification DTO")
	}
	result := map[string]any{}
	result["id"] = value.Id
	result["recipient"] = value.Recipient
	result["subject"] = value.Subject
	result["body"] = value.Body
	result["status"] = value.Status
	if value.ProviderReceipt != nil {
		result["providerReceipt"] = *value.ProviderReceipt
	}
	return result, nil
}
func mapcommunicationLoadedNotification(value *communication.LoadedNotification) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing LoadedNotification DTO")
	}
	result := map[string]any{}
	result["exists"] = value.Exists
	result["version"] = value.Version
	itemState, err := mapcommunicationNotification(value.State)
	if err != nil {
		return nil, err
	}
	result["state"] = itemState
	if !value.Exists || value.State == nil {
		return nil, fmt.Errorf("invalid loaded DTO")
	}
	return result, nil
}
func mapcommunicationNotifications(value *communication.Notifications) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Notifications DTO")
	}
	result := map[string]any{}
	itemsItems := make([]any, 0, len(value.Items))
	for _, child := range value.Items {
		item, err := mapcommunicationLoadedNotification(child)
		if err != nil {
			return nil, err
		}
		itemsItems = append(itemsItems, item)
	}
	result["items"] = itemsItems
	return result, nil
}
func maployaltyGrant(value *loyalty.Grant) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Grant DTO")
	}
	result := map[string]any{}
	result["id"] = value.Id
	result["accountId"] = value.AccountId
	result["benefit"] = value.Benefit
	result["validDays"] = value.ValidDays
	return result, nil
}
func maployaltyAccount(value *loyalty.Account) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Account DTO")
	}
	result := map[string]any{}
	result["id"] = value.Id
	result["stampBalance"] = value.StampBalance
	result["collections"] = value.Collections
	result["grantsEarned"] = value.GrantsEarned
	if value.LastGrant != nil {
		itemLastGrant, err := maployaltyGrant(value.LastGrant)
		if err != nil {
			return nil, err
		}
		result["lastGrant"] = itemLastGrant
	}
	return result, nil
}
func maployaltyReward(value *loyalty.Reward) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Reward DTO")
	}
	result := map[string]any{}
	result["id"] = value.Id
	result["grantId"] = value.GrantId
	result["customerId"] = value.CustomerId
	result["benefit"] = value.Benefit
	result["status"] = value.Status
	result["expiresAt"] = value.ExpiresAt
	if value.RedeemedFor != nil {
		result["redeemedFor"] = *value.RedeemedFor
	}
	return result, nil
}
func maployaltyLoadedAccount(value *loyalty.LoadedAccount) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing LoadedAccount DTO")
	}
	result := map[string]any{}
	result["exists"] = value.Exists
	result["version"] = value.Version
	itemState, err := maployaltyAccount(value.State)
	if err != nil {
		return nil, err
	}
	result["state"] = itemState
	if !value.Exists || value.State == nil {
		return nil, fmt.Errorf("invalid loaded DTO")
	}
	return result, nil
}
func maployaltyAccounts(value *loyalty.Accounts) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Accounts DTO")
	}
	result := map[string]any{}
	itemsItems := make([]any, 0, len(value.Items))
	for _, child := range value.Items {
		item, err := maployaltyLoadedAccount(child)
		if err != nil {
			return nil, err
		}
		itemsItems = append(itemsItems, item)
	}
	result["items"] = itemsItems
	return result, nil
}
func maployaltyLoadedReward(value *loyalty.LoadedReward) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing LoadedReward DTO")
	}
	result := map[string]any{}
	result["exists"] = value.Exists
	result["version"] = value.Version
	itemState, err := maployaltyReward(value.State)
	if err != nil {
		return nil, err
	}
	result["state"] = itemState
	if !value.Exists || value.State == nil {
		return nil, fmt.Errorf("invalid loaded DTO")
	}
	return result, nil
}
func maployaltyRewards(value *loyalty.Rewards) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Rewards DTO")
	}
	result := map[string]any{}
	itemsItems := make([]any, 0, len(value.Items))
	for _, child := range value.Items {
		item, err := maployaltyLoadedReward(child)
		if err != nil {
			return nil, err
		}
		itemsItems = append(itemsItems, item)
	}
	result["items"] = itemsItems
	return result, nil
}
func mapmenuDrink(value *menu.Drink) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Drink DTO")
	}
	result := map[string]any{}
	result["id"] = value.Id
	result["name"] = value.Name
	result["revision"] = value.Revision
	result["published"] = value.Published
	return result, nil
}
func mapmenuOffer(value *menu.Offer) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Offer DTO")
	}
	result := map[string]any{}
	result["code"] = value.Code
	result["drinkId"] = value.DrinkId
	result["drinkRevision"] = value.DrinkRevision
	result["name"] = value.Name
	result["minor"] = value.Minor
	result["currency"] = value.Currency
	return result, nil
}
func mapmenuEdition(value *menu.Edition) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Edition DTO")
	}
	result := map[string]any{}
	result["id"] = value.Id
	result["currency"] = value.Currency
	result["status"] = value.Status
	itemsOffers := make([]any, 0, len(value.Offers))
	for _, child := range value.Offers {
		item, err := mapmenuOffer(child)
		if err != nil {
			return nil, err
		}
		itemsOffers = append(itemsOffers, item)
	}
	result["offers"] = itemsOffers
	return result, nil
}
func mapmenuLoadedDrink(value *menu.LoadedDrink) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing LoadedDrink DTO")
	}
	result := map[string]any{}
	result["exists"] = value.Exists
	result["version"] = value.Version
	itemState, err := mapmenuDrink(value.State)
	if err != nil {
		return nil, err
	}
	result["state"] = itemState
	if !value.Exists || value.State == nil {
		return nil, fmt.Errorf("invalid loaded DTO")
	}
	return result, nil
}
func mapmenuDrinks(value *menu.Drinks) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Drinks DTO")
	}
	result := map[string]any{}
	itemsItems := make([]any, 0, len(value.Items))
	for _, child := range value.Items {
		item, err := mapmenuLoadedDrink(child)
		if err != nil {
			return nil, err
		}
		itemsItems = append(itemsItems, item)
	}
	result["items"] = itemsItems
	return result, nil
}
func mapmenuLoadedEdition(value *menu.LoadedEdition) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing LoadedEdition DTO")
	}
	result := map[string]any{}
	result["exists"] = value.Exists
	result["version"] = value.Version
	itemState, err := mapmenuEdition(value.State)
	if err != nil {
		return nil, err
	}
	result["state"] = itemState
	if !value.Exists || value.State == nil {
		return nil, fmt.Errorf("invalid loaded DTO")
	}
	return result, nil
}
func mapmenuEditions(value *menu.Editions) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Editions DTO")
	}
	result := map[string]any{}
	itemsItems := make([]any, 0, len(value.Items))
	for _, child := range value.Items {
		item, err := mapmenuLoadedEdition(child)
		if err != nil {
			return nil, err
		}
		itemsItems = append(itemsItems, item)
	}
	result["items"] = itemsItems
	return result, nil
}
func maporderingSelection(value *ordering.Selection) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Selection DTO")
	}
	result := map[string]any{}
	result["offerCode"] = value.OfferCode
	result["name"] = value.Name
	result["minor"] = value.Minor
	return result, nil
}
func maporderingLine(value *ordering.Line) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Line DTO")
	}
	result := map[string]any{}
	result["id"] = value.Id
	itemSelection, err := maporderingSelection(value.Selection)
	if err != nil {
		return nil, err
	}
	result["selection"] = itemSelection
	result["quantity"] = value.Quantity
	return result, nil
}
func maporderingOrder(value *ordering.Order) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Order DTO")
	}
	result := map[string]any{}
	result["id"] = value.Id
	result["customerId"] = value.CustomerId
	result["editionId"] = value.EditionId
	result["currency"] = value.Currency
	result["status"] = value.Status
	itemsLines := make([]any, 0, len(value.Lines))
	for _, child := range value.Lines {
		item, err := maporderingLine(child)
		if err != nil {
			return nil, err
		}
		itemsLines = append(itemsLines, item)
	}
	result["lines"] = itemsLines
	return result, nil
}
func maporderingLoadedOrder(value *ordering.LoadedOrder) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing LoadedOrder DTO")
	}
	result := map[string]any{}
	result["exists"] = value.Exists
	result["version"] = value.Version
	itemState, err := maporderingOrder(value.State)
	if err != nil {
		return nil, err
	}
	result["state"] = itemState
	if !value.Exists || value.State == nil {
		return nil, fmt.Errorf("invalid loaded DTO")
	}
	return result, nil
}
func maporderingOrders(value *ordering.Orders) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Orders DTO")
	}
	result := map[string]any{}
	itemsItems := make([]any, 0, len(value.Items))
	for _, child := range value.Items {
		item, err := maporderingLoadedOrder(child)
		if err != nil {
			return nil, err
		}
		itemsItems = append(itemsItems, item)
	}
	result["items"] = itemsItems
	return result, nil
}
func mappreparationTicket(value *preparation.Ticket) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Ticket DTO")
	}
	result := map[string]any{}
	result["id"] = value.Id
	result["orderId"] = value.OrderId
	result["customerId"] = value.CustomerId
	result["instructions"] = value.Instructions
	result["status"] = value.Status
	return result, nil
}
func mappreparationLoadedTicket(value *preparation.LoadedTicket) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing LoadedTicket DTO")
	}
	result := map[string]any{}
	result["exists"] = value.Exists
	result["version"] = value.Version
	itemState, err := mappreparationTicket(value.State)
	if err != nil {
		return nil, err
	}
	result["state"] = itemState
	if !value.Exists || value.State == nil {
		return nil, fmt.Errorf("invalid loaded DTO")
	}
	return result, nil
}
func mappreparationTickets(value *preparation.Tickets) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("missing Tickets DTO")
	}
	result := map[string]any{}
	itemsItems := make([]any, 0, len(value.Items))
	for _, child := range value.Items {
		item, err := mappreparationLoadedTicket(child)
		if err != nil {
			return nil, err
		}
		itemsItems = append(itemsItems, item)
	}
	result["items"] = itemsItems
	return result, nil
}
