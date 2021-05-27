package main

func main() {
	//conn, err := grpc.Dial("localhost:4001", grpc.WithInsecure())
	//if err != nil {
	//	logrus.Fatalln("Failed making connection to shard")
	//}
	//defer conn.Close()
	//
	//shardClient := pb.NewShardClient(conn)
	//
	//time.Sleep(3 * time.Second)
	//for range time.Tick(1 * time.Second) {
	//	start := time.Now()
	//	record, err := shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo"})
	//	appendTime := time.Since(start)
	//	if err != nil {
	//		logrus.Fatalln(err)
	//		return
	//	}
	//	log.Printf("Received %v with GSN %v; Time: %v", record.Record, record.Gsn, appendTime)
	//}
}
