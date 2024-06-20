package external

import (
	"context"
	booking "github.com/ZMS-DevOps/booking-service/proto"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/mmmajder/zms-devops-grade-service/util"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func NewBookingClient(address string) booking.BookingServiceClient {
	conn, err := getConnection(address)
	if err != nil {
		log.Fatalf("Failed to start gRPC connection to Catalogue service: %v", err)
	}
	return booking.NewBookingServiceClient(conn)
}

func getConnection(address string) (*grpc.ClientConn, error) {
	return grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func IfHostCanBeDeleted(bookingClient booking.BookingServiceClient, id string, span trace.Span, loki promtail.Client) (*booking.CheckDeleteHostResponse, error) {
	util.HttpTraceInfo("Checking if host can be deleted...", span, loki, "IfHostCanBeDeleted", "")
	return bookingClient.CheckDeleteHost(
		context.TODO(),
		&booking.CheckDeleteHostRequest{
			HostId: id,
		})
}

func IfGuestCanBeDeleted(bookingClient booking.BookingServiceClient, id string, span trace.Span, loki promtail.Client) (*booking.CheckDeleteClientResponse, error) {
	util.HttpTraceInfo("Checking if guest can be deleted...", span, loki, "IfGuestCanBeDeleted", "")
	return bookingClient.CheckDeleteClient(
		context.TODO(),
		&booking.CheckDeleteClientRequest{
			HostId: id,
		})
}

func IfGuestCanReviewHost(bookingClient booking.BookingServiceClient, reviewerSub string, reviewedSub string, span trace.Span, loki promtail.Client) (*booking.CheckGuestHasReservationForHostResponse, error) {
	util.HttpTraceInfo("Checking if guest can review host...", span, loki, "IfGuestCanReviewHost", "")
	return bookingClient.CheckGuestHasReservationForHost(
		context.TODO(),
		&booking.CheckGuestHasReservationForHostRequest{
			ReviewerId: reviewerSub,
			HostId:     reviewedSub,
		})
}

func IfGuestCanReviewAccommodation(bookingClient booking.BookingServiceClient, reviewerSub string, reviewedSub string, span trace.Span, loki promtail.Client) (*booking.CheckGuestHasReservationForAccommodationResponse, error) {
	util.HttpTraceInfo("Checking if guest can review accommodation...", span, loki, "IfGuestCanReviewAccommodation", "")
	return bookingClient.CheckGuestHasReservationForAccommodation(
		context.TODO(),
		&booking.CheckGuestHasReservationForAccommodationRequest{
			ReviewerId:      reviewerSub,
			AccommodationId: reviewedSub,
		})
}
