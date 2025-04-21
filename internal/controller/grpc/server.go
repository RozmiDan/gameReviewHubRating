package rating_server

import (
	"context"

	"github.com/RozmiDan/gameReviewHubRating/internal/entity"
	ratingv1 "github.com/RozmiDan/gamehub-protos/gen/go/gamehub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RatingService interface {
	SubmitRating(ctx context.Context, userID string, gameID string, rating int32) error
	GetGameRating(ctx context.Context, gameID string) (entity.GameRating, error)
	GetTopGames(ctx context.Context, limit, offset int32) ([]entity.GameRating, error)
}

type serverAPI struct {
	ratingv1.UnimplementedRatingServiceServer
	service RatingService
}

func Register(grpcServer *grpc.Server, svc RatingService) {
	ratingv1.RegisterRatingServiceServer(grpcServer, &serverAPI{service: svc})
}

func (s *serverAPI) SubmitRating(ctx context.Context,
	req *ratingv1.SubmitRatingRequest) (*ratingv1.SubmitRatingResponse, error) {

	if req.GetGameId() == "" || req.GetUserId() == "" || req.GetRating() == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid entered data")
	}

	if err := s.service.SubmitRating(ctx, req.UserId, req.GameId, req.Rating); err != nil {
		return &ratingv1.SubmitRatingResponse{Success: false}, status.Error(codes.Internal, "could not submit rating")
	}

	return &ratingv1.SubmitRatingResponse{Success: true}, nil

}

func (s *serverAPI) GetGameRating(ctx context.Context,
	req *ratingv1.GetGameRatingRequest) (*ratingv1.GetGameRatingResponse, error) {

	if req.GetGameId() == "" {
		return &ratingv1.GetGameRatingResponse{
			GameId:        "",
			AverageRating: 0,
			RatingsCount:  0,
		}, status.Error(codes.InvalidArgument, "invalid gameID")
	}

	resEnt, err := s.service.GetGameRating(ctx, req.GameId)

	if err != nil {
		return &ratingv1.GetGameRatingResponse{
			GameId:        "",
			AverageRating: 0,
			RatingsCount:  0,
		}, status.Error(codes.NotFound, "rating not found")
	}
	return &ratingv1.GetGameRatingResponse{
		GameId:        resEnt.GameId,
		AverageRating: float64(resEnt.AverageRating),
		RatingsCount:  resEnt.RatingsCount,
	}, nil
}

func (s *serverAPI) GetTopGames(ctx context.Context,
	req *ratingv1.GetTopGamesRequest) (*ratingv1.GetTopGamesResponse, error) {

	if req.GetLimit() != 10 {
		return &ratingv1.GetTopGamesResponse{}, status.Error(codes.InvalidArgument, "invalid limit")
	}

	list, err := s.service.GetTopGames(ctx, req.Limit, req.Offset)

	if err != nil {
		return nil, status.Error(codes.Internal, "could not get top games")
	}

	resp := &ratingv1.GetTopGamesResponse{}
	for _, e := range list {
		resp.Games = append(resp.Games, &ratingv1.GameRating{
			GameId:        e.GameId,
			AverageRating: float64(e.AverageRating),
			RatingsCount:  e.RatingsCount,
		})
	}
	return resp, nil
}
