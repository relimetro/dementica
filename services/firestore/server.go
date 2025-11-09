package main

import(
	"log" // for loggin
	"sync" // for mutex
	"google.golang.org/api/iterator"

	// Grpc
	"net"
	pb "example/proto_example/protoOut"
	aiProompt "example/proto_example/protoAI"
	"google.golang.org/grpc"
	"golang.org/x/net/context"

	// firebase
	firestore "cloud.google.com/go/firestore"
	"example/proto_example/myFire"
)



///////////////////////////////////////////////////////////////
/// Functionality

func (s *server) ProcessLifestyle(x string) (string,bool) {

	var conn *grpc.ClientConn
	// conn, err := grpc.Dial("vertexai:50052", grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil { log.Printf("[ERROR] GRPC: cound not connect vertexAI at 50052: \n%s",err); return "",false; }
	defer conn.Close()
	c := aiProompt.NewAiProomptClient(conn)

	txt := "Diabetic:true,AlcoholLevel:0.084973629, HeartRate:98, BloodOxygenLevel:96.23074296, BodyTemperature:36.22485168, Weight:57.56397754, MRI_Delay:36.42102798, Presecription:None, DosageMg:0, Age:60, EducationLevel:Primary School, DominantHand:Left, Gender:Female, FamilyHistory:false, SmokingStatus:Current Smoker, APOE_e19:false, PhysicalActivity:Sedentary, DepressionStatus:false, MedicationHistory:false, NutritionDiet:Low-Carb Diet, SleepQuality:Poor, ChronicHealthConditionsDiabetes"
	// txt = "short response why is the sky blue"
	message := aiProompt.ProomptMsg { Message: txt}
	resp, err := c.HealtcareProompt(context.Background(), &message)
	if err != nil { log.Printf("[ERROR] FTproompt, <%s>, <%d>",err,resp); return "",false; }
	log.Printf("Response FTproompt: %s",resp.Message)
	return resp.Message,true }






///////////////////////////////////////////////////////////////
/// GRPC
type server struct{
	pb.UnimplementedFirestoreServer
	client *firestore.Client
}



// Login (UserLogin -> SessionToken)
func (s *server) Login(ctx context.Context, x *pb.UserLogin) (*pb.SessionToken, error) {
	log.Printf("login: %s, %s\n\n",x.UserName, x.PlaintextPassword)

	succ, UserId, _ := myFire.Login(s.c,x.Username,x.PlaintextPassword)

	// return session token
	return &pb.SessionToken{
		Temp: UserId,
	}, nil
}



// GetDetails (UserRequest -> UserDetails)
func (s *server) GetDetails(ctx context.Context, x *pb.UserRequest) (*pb.UserDetails, error) {
	log.Printf("GetDetails: %s, %s", x.SessionToken, x.UserId)
	return &pb.UserDetails{
		Details:"DEPRICATED" }, nil
}



// GetRisk (SessionToken -> RiskScore)
func (s *server) GetRisk(ctx context.Context, x *pb.SessionToken) (*pb.RiskScore, error) {

	p := myFire.GetPatientInfo(s.c, )

	// Dummy Response
	return &pb.RiskScore{ Score: 0, }, nil
}



// SendLifestyle (SessionToken -> RiskScore)
func (s *server) SendLifestyle(ctx context.Context, x *pb.LifestyleRequest) (*pb.LifestyleResponse, error) {
	log.Printf("SendLifestyle:'%s'", x.Message)

	calc_risk, ok := s.ProcessLifestyle(x.Message) // keras
	if ok == false { calc_risk = "Error Calculating"; }
	log.Printf("calc_risk: %s\n",calc_risk)
	// TODO: log errors that occur in database, or some log file?

	_, _, err2 := s.client.Collection("patientData").Add(ctx, map[string]interface{}{
		"data":x.Message,
		"calculated_risk":calc_risk,
	})
	if err2 != nil { log.Fatalf("Failed adding\n%v", err2)}

	// Dummy Response
	return &pb.LifestyleResponse{ Success: true, }, nil
}






///////////////////////////////////////////////////////////////
/// Main

func main() {
	// grpc connection
	lis, err := net.Listen("tcp", ":9000");
	if err != nil { log.Fatalf("GRPC: failed to listen:\n%v", err) }

	// serv GRPC
	serverData := server{client:myFire.FirebaseInit()}
	defer serverData.client.Close()

	// Reset firestore
	myFire.BURN_IT_ALL_DOWN(serverData.client)


	// start server
	grpcServer := grpc.NewServer()
	pb.RegisterFirestoreServer(grpcServer, &serverData)
	log.Printf("Ready!! >:0")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("GRPC: Failed to serve:\n%v",err) }
}
