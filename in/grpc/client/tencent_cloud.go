package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"interview/common/global"
	proto2 "interview/proto"
	"time"
)

func CreateRecTask(audioUrl, callbackUrl string) (uint64, error) {
	conn, err := grpc.Dial(global.CONFIG.ServiceUrls.GPTServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	requestParams := proto2.TencentCloudCreateRecRequest{
		AudioUrl:    audioUrl,
		CallbackUrl: callbackUrl,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	client := proto2.NewTencentCloudClient(conn)
	info, err := client.CreateRecTask(ctx, &requestParams)
	if err != nil {
		return 0, err
	}
	if info.Code != 200 {
		return 0, fmt.Errorf(info.Msg)
	}
	return info.Data, nil
}
