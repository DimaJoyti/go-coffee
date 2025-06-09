package grpc

import (
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	pb "github.com/DimaJoyti/go-coffee/proto/kitchen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Equipment Converters

func convertStationType(pbType pb.StationType) domain.StationType {
	switch pbType {
	case pb.StationType_STATION_TYPE_ESPRESSO:
		return domain.StationTypeEspresso
	case pb.StationType_STATION_TYPE_GRINDER:
		return domain.StationTypeGrinder
	case pb.StationType_STATION_TYPE_STEAMER:
		return domain.StationTypeSteamer
	case pb.StationType_STATION_TYPE_ASSEMBLY:
		return domain.StationTypeAssembly
	default:
		return domain.StationTypeUnknown
	}
}

func convertStationTypeToProto(domainType domain.StationType) pb.StationType {
	switch domainType {
	case domain.StationTypeEspresso:
		return pb.StationType_STATION_TYPE_ESPRESSO
	case domain.StationTypeGrinder:
		return pb.StationType_STATION_TYPE_GRINDER
	case domain.StationTypeSteamer:
		return pb.StationType_STATION_TYPE_STEAMER
	case domain.StationTypeAssembly:
		return pb.StationType_STATION_TYPE_ASSEMBLY
	default:
		return pb.StationType_STATION_TYPE_UNSPECIFIED
	}
}

func convertEquipmentStatus(pbStatus pb.EquipmentStatus) domain.EquipmentStatus {
	switch pbStatus {
	case pb.EquipmentStatus_EQUIPMENT_STATUS_AVAILABLE:
		return domain.EquipmentStatusAvailable
	case pb.EquipmentStatus_EQUIPMENT_STATUS_IN_USE:
		return domain.EquipmentStatusInUse
	case pb.EquipmentStatus_EQUIPMENT_STATUS_MAINTENANCE:
		return domain.EquipmentStatusMaintenance
	case pb.EquipmentStatus_EQUIPMENT_STATUS_BROKEN:
		return domain.EquipmentStatusBroken
	default:
		return domain.EquipmentStatusUnknown
	}
}

func convertEquipmentStatusToProto(domainStatus domain.EquipmentStatus) pb.EquipmentStatus {
	switch domainStatus {
	case domain.EquipmentStatusAvailable:
		return pb.EquipmentStatus_EQUIPMENT_STATUS_AVAILABLE
	case domain.EquipmentStatusInUse:
		return pb.EquipmentStatus_EQUIPMENT_STATUS_IN_USE
	case domain.EquipmentStatusMaintenance:
		return pb.EquipmentStatus_EQUIPMENT_STATUS_MAINTENANCE
	case domain.EquipmentStatusBroken:
		return pb.EquipmentStatus_EQUIPMENT_STATUS_BROKEN
	default:
		return pb.EquipmentStatus_EQUIPMENT_STATUS_UNSPECIFIED
	}
}

func convertEquipmentToProto(equipment *application.EquipmentResponse) *pb.EquipmentResponse {
	return &pb.EquipmentResponse{
		Id:               equipment.ID,
		Name:             equipment.Name,
		StationType:      convertStationTypeToProto(equipment.StationType),
		Status:           convertEquipmentStatusToProto(equipment.Status),
		EfficiencyScore:  equipment.EfficiencyScore,
		CurrentLoad:      equipment.CurrentLoad,
		MaxCapacity:      equipment.MaxCapacity,
		UtilizationRate:  equipment.UtilizationRate,
		IsAvailable:      equipment.IsAvailable,
		NeedsMaintenance: equipment.NeedsMaintenance,
		LastMaintenance:  timestamppb.New(equipment.LastMaintenance),
		CreatedAt:        timestamppb.New(equipment.CreatedAt),
		UpdatedAt:        timestamppb.New(equipment.UpdatedAt),
	}
}

// Staff Converters

func convertStaffToProto(staff *application.StaffResponse) *pb.StaffResponse {
	specializations := make([]pb.StationType, len(staff.Specializations))
	for i, spec := range staff.Specializations {
		specializations[i] = convertStationTypeToProto(spec)
	}

	return &pb.StaffResponse{
		Id:                  staff.ID,
		Name:                staff.Name,
		Specializations:     specializations,
		SkillLevel:          staff.SkillLevel,
		IsAvailable:         staff.IsAvailable,
		CurrentOrders:       staff.CurrentOrders,
		MaxConcurrentOrders: staff.MaxConcurrentOrders,
		Workload:            staff.Workload,
		IsOverloaded:        staff.IsOverloaded,
		CreatedAt:           timestamppb.New(staff.CreatedAt),
		UpdatedAt:           timestamppb.New(staff.UpdatedAt),
	}
}

// Order Converters

func convertOrderPriority(pbPriority pb.OrderPriority) domain.OrderPriority {
	switch pbPriority {
	case pb.OrderPriority_ORDER_PRIORITY_LOW:
		return domain.OrderPriorityLow
	case pb.OrderPriority_ORDER_PRIORITY_NORMAL:
		return domain.OrderPriorityNormal
	case pb.OrderPriority_ORDER_PRIORITY_HIGH:
		return domain.OrderPriorityHigh
	case pb.OrderPriority_ORDER_PRIORITY_URGENT:
		return domain.OrderPriorityUrgent
	default:
		return domain.OrderPriorityNormal
	}
}

func convertOrderPriorityToProto(domainPriority domain.OrderPriority) pb.OrderPriority {
	switch domainPriority {
	case domain.OrderPriorityLow:
		return pb.OrderPriority_ORDER_PRIORITY_LOW
	case domain.OrderPriorityNormal:
		return pb.OrderPriority_ORDER_PRIORITY_NORMAL
	case domain.OrderPriorityHigh:
		return pb.OrderPriority_ORDER_PRIORITY_HIGH
	case domain.OrderPriorityUrgent:
		return pb.OrderPriority_ORDER_PRIORITY_URGENT
	default:
		return pb.OrderPriority_ORDER_PRIORITY_NORMAL
	}
}

func convertOrderStatus(pbStatus pb.OrderStatus) domain.OrderStatus {
	switch pbStatus {
	case pb.OrderStatus_ORDER_STATUS_PENDING:
		return domain.OrderStatusPending
	case pb.OrderStatus_ORDER_STATUS_PROCESSING:
		return domain.OrderStatusProcessing
	case pb.OrderStatus_ORDER_STATUS_COMPLETED:
		return domain.OrderStatusCompleted
	case pb.OrderStatus_ORDER_STATUS_CANCELLED:
		return domain.OrderStatusCancelled
	default:
		return domain.OrderStatusUnknown
	}
}

func convertOrderStatusToProto(domainStatus domain.OrderStatus) pb.OrderStatus {
	switch domainStatus {
	case domain.OrderStatusPending:
		return pb.OrderStatus_ORDER_STATUS_PENDING
	case domain.OrderStatusProcessing:
		return pb.OrderStatus_ORDER_STATUS_PROCESSING
	case domain.OrderStatusCompleted:
		return pb.OrderStatus_ORDER_STATUS_COMPLETED
	case domain.OrderStatusCancelled:
		return pb.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return pb.OrderStatus_ORDER_STATUS_UNKNOWN
	}
}

func convertOrderToProto(order *application.OrderResponse) *pb.OrderResponse {
	items := make([]*pb.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		requirements := make([]pb.StationType, len(item.Requirements))
		for j, req := range item.Requirements {
			requirements[j] = convertStationTypeToProto(req)
		}

		items[i] = &pb.OrderItemResponse{
			Id:           item.ID,
			Name:         item.Name,
			Quantity:     item.Quantity,
			Instructions: item.Instructions,
			Requirements: requirements,
			Metadata:     item.Metadata,
		}
	}

	requiredStations := make([]pb.StationType, len(order.RequiredStations))
	for i, station := range order.RequiredStations {
		requiredStations[i] = convertStationTypeToProto(station)
	}

	response := &pb.OrderResponse{
		Id:                  order.ID,
		CustomerId:          order.CustomerID,
		Items:               items,
		Status:              convertOrderStatusToProto(order.Status),
		Priority:            convertOrderPriorityToProto(order.Priority),
		EstimatedTime:       order.EstimatedTime,
		ActualTime:          order.ActualTime,
		AssignedStaffId:     order.AssignedStaffID,
		AssignedEquipment:   order.AssignedEquipment,
		SpecialInstructions: order.SpecialInstructions,
		RequiredStations:    requiredStations,
		TotalQuantity:       order.TotalQuantity,
		WaitTime:            order.WaitTime,
		ProcessingTime:      order.ProcessingTime,
		IsOverdue:           order.IsOverdue,
		IsReadyToStart:      order.IsReadyToStart,
		CreatedAt:           timestamppb.New(order.CreatedAt),
		UpdatedAt:           timestamppb.New(order.UpdatedAt),
	}

	if order.StartedAt != nil {
		response.StartedAt = timestamppb.New(*order.StartedAt)
	}

	if order.CompletedAt != nil {
		response.CompletedAt = timestamppb.New(*order.CompletedAt)
	}

	return response
}

// Queue Converters

func convertQueueStatusToProto(status *application.QueueStatusResponse) *pb.QueueStatusResponse {
	queuesByPriority := make(map[string]int32)
	for priority, count := range status.QueuesByPriority {
		queuesByPriority[convertOrderPriorityToProto(priority).String()] = count
	}

	stationLoad := make(map[string]float32)
	for station, load := range status.StationLoad {
		stationLoad[convertStationTypeToProto(station).String()] = load
	}

	overdueOrders := make([]*pb.OrderResponse, len(status.OverdueOrders))
	for i, order := range status.OverdueOrders {
		overdueOrders[i] = convertOrderToProto(order)
	}

	response := &pb.QueueStatusResponse{
		TotalOrders:      status.TotalOrders,
		ProcessingOrders: status.ProcessingOrders,
		PendingOrders:    status.PendingOrders,
		CompletedOrders:  status.CompletedOrders,
		AverageWaitTime:  status.AverageWaitTime,
		QueuesByPriority: queuesByPriority,
		StationLoad:      stationLoad,
		OverdueOrders:    overdueOrders,
		UpdatedAt:        timestamppb.New(status.UpdatedAt),
	}

	if status.NextOrder != nil {
		response.NextOrder = convertOrderToProto(status.NextOrder)
	}

	return response
}

// Optimization Converters

func convertOptimizationToProto(optimization *application.OptimizationResponse) *pb.OptimizeQueueResponse {
	steps := make([]*pb.WorkflowStepResponse, len(optimization.OptimizedSteps))
	for i, step := range optimization.OptimizedSteps {
		steps[i] = &pb.WorkflowStepResponse{
			StepId:         step.StepID,
			StationType:    convertStationTypeToProto(step.StationType),
			EstimatedTime:  step.EstimatedTime,
			RequiredSkill:  step.RequiredSkill,
			Dependencies:   step.Dependencies,
			CanParallelize: step.CanParallelize,
			EquipmentId:    step.EquipmentID,
			StaffId:        step.StaffID,
		}
	}

	staffAllocations := make([]*pb.StaffAllocationResponse, len(optimization.StaffAllocations))
	for i, allocation := range optimization.StaffAllocations {
		staffAllocations[i] = &pb.StaffAllocationResponse{
			StaffId:          allocation.StaffID,
			OrderId:          allocation.OrderID,
			StationType:      convertStationTypeToProto(allocation.StationType),
			EstimatedTime:    allocation.EstimatedTime,
			EfficiencyScore:  allocation.EfficiencyScore,
			AllocationReason: allocation.AllocationReason,
			AllocatedAt:      timestamppb.New(allocation.AllocatedAt),
		}
	}

	equipmentAllocations := make([]*pb.EquipmentAllocationResponse, len(optimization.EquipmentAllocations))
	for i, allocation := range optimization.EquipmentAllocations {
		equipmentAllocations[i] = &pb.EquipmentAllocationResponse{
			EquipmentId: allocation.EquipmentID,
			OrderIds:    allocation.OrderIDs,
			StartTime:   timestamppb.New(allocation.StartTime),
			EndTime:     timestamppb.New(allocation.EndTime),
			Load:        allocation.Load,
		}
	}

	return &pb.OptimizeQueueResponse{
		OrderId:              optimization.OrderID,
		OptimizedSteps:       steps,
		EstimatedTime:        optimization.EstimatedTime,
		EfficiencyGain:       optimization.EfficiencyGain,
		ResourceUtilization:  optimization.ResourceUtilization,
		Recommendations:      optimization.Recommendations,
		StaffAllocations:     staffAllocations,
		EquipmentAllocations: equipmentAllocations,
		CreatedAt:            timestamppb.New(optimization.CreatedAt),
	}
}

// Helper function to convert bool pointer for protobuf
func boolValue(b bool) *wrapperspb.BoolValue {
	return &wrapperspb.BoolValue{Value: b}
}
