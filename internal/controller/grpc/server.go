package rating_server

import (
	"context"

	ratingv1 "github.com/RozmiDan/gamehub-protos/gen/go/gamehub"
	"google.golang.org/grpc"
)

type serverAPI struct {
	ratingv1.UnimplementedRatingServiceServer
}

func Register(gRPC *grpc.Server) {
	ratingv1.RegisterRatingServiceServer(gRPC, &serverAPI{})
}

func (s *serverAPI) SubmitRating(ctx context.Context,
	req *ratingv1.SubmitRatingRequest) (*ratingv1.SubmitRatingResponse, error) {
	panic("implement me")
}

func (s *serverAPI) GetGameRating(ctx context.Context,
	req *ratingv1.GetGameRatingRequest) (*ratingv1.GetGameRatingResponse, error) {
	panic("implement me")
}

func (s *serverAPI) GetTopGames(ctx context.Context,
	req *ratingv1.GetTopGamesRequest) (*ratingv1.GetTopGamesResponse, error) {
	panic("implement me")
}
