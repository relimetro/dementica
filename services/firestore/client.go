package main

import (
	"log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "example/proto_example/protoOut"
)

func main() {
	// connection
	var conn *grpc.ClientConn
	// conn, err := grpc.Dial(":9000", grpc.WithInsecure() )
	conn, err := grpc.Dial(":9000", grpc.WithInsecure() )
	if err != nil { log.Fatalf("GRPC: could not connect,\n%s", err)}
	defer conn.Close()
	c := pb.NewFirestoreClient(conn)

	// Login
	message := pb.UserLogin { UserName: "name", PlaintextPassword:"pass", }
	mySession, err := c.Login(context.Background(), &message)
	if err != nil { log.Fatalf("Err: send msg: %s", err) }
	log.Printf("Response from server: %d", mySession.Temp)

	// Get Details
	message2 := pb.UserRequest { UserId: "name", SessionToken:mySession.Temp, }
	response2, err := c.GetDetails(context.Background(), &message2)
	if err != nil { log.Fatalf("Err: send msg2: %s", err) }
	log.Printf("Response from server: %s", response2.Details)

	// Invalid GetDetails
	message3 := pb.UserRequest { UserId: "nam", SessionToken:mySession.Temp, }
	response3, err := c.GetDetails(context.Background(), &message3)
	if err != nil { log.Fatalf("Err: send msg3: %s", err) }
	log.Printf("Response from server: %s", response3.Details)

	// Get Risk
	message4 := *mySession
	response4, err := c.GetRisk(context.Background(), &message4)
	if err != nil { log.Fatalf("Err: send msg4: %s", err) }
	log.Printf("Response from server: %d", response4.Score)

	// Send lifestyle
	lifestyle := "Diabetic:true,AlcoholLevel:0.084973629, HeartRate:98, BloodOxygenLevel:96.23074296, BodyTemperature:36.22485168, Weight:57.56397754, MRI_Delay:36.42102798, Presecription:None, DosageMg:0, Age:60, EducationLevel:Primary School, DominantHand:Left, Gender:Female, FamilyHistory:false, SmokingStatus:Current Smoker, APOE_e19:false, PhysicalActivity:Sedentary, DepressionStatus:false, MedicationHistory:false, NutritionDiet:Low-Carb Diet, SleepQuality:Poor, ChronicHealthConditionsDiabetes"
	message5 := pb.LifestyleRequest { Message:lifestyle}
	response5, err := c.SendLifestyle(context.Background(), &message5)
	if err != nil { log.Fatalf("Err: send msg5: %s", err) }
	log.Printf("lifestyle from server: %d", response5.Success)

}
