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
	c *firestore.Client
}



// Register
func (s *server) Register(ctx context.Context, x *pb.UserRegister) (*pb.RegisterResult, error) {
	log.Printf("register:") //%s, %s\n\n",x.UserName, x.PlaintextPassword)

	if x.UserType == pb.UserType_Patient {
		myFire.RegisterPatient(s.c, x.Name, x.Password, "") // todo ability to assign to doctor
	} else {
		myFire.RegisterDoctor(s.c, x.Name, x.Password, x.RegisterWith) // todo check for correct email etc
	}


	// return session token
	return &pb.RegisterResult{
		Result: pb.RegResult_Ok,
	}, nil
}

// Login (UserLogin -> LoginResult)
func (s *server) Login(ctx context.Context, x *pb.UserLogin) (*pb.LoginResult, error) {
	log.Printf("login: %s, %s\n\n",x.Name, x.Password)

	_, UserId, _ := myFire.Login(s.c,x.Name,x.Password)
	// succ, UserId, _ := myFire.Login(s.c,x.Username,x.PlaintextPassword)

	// return session token
	return &pb.LoginResult{
		UserID: UserId,
		Result: RegResult.Ok,
	}, nil
}


// PatientInfo
func (s *server) PatientInfo(ctx context.Context, x *pb.UserID) (*pb.PatientData, error) {
	log.Printf("patientInfo:")

	p, _ := myFire.GetPatientInfo(s.c, x.UserID)

	// return session token
	return &pb.PatientData{
		Result: pb.PatientData_Ok,
		Name: p.Name,
		HasDementia: p.HasDementia,
		DocotorID: p.DoctorID,
		RiskScore: p.RiskScore,
	}, nil
}
// Get Risk
func (s *server) GetRisk(ctx context.Context, x *pb.UserID) (*pb.RiskResponse, error) {
	log.Printf("GetRisk:")

	p, _ := myFire.GetPatientInfo(s.c, x.UserID)

	// return session token
	return &pb.RiskResponse{
		Result: pb.RiskResponse_Ok,
		RiskScore: p.RiskScore,
	}, nil
}


// Doctor Info
func (s *server) DoctorInfo(ctx context.Context, x *pb.UserID) (*pb.DoctorData, error) {
	log.Printf("DoctorInfo:")

	d, _ := myFire.GetDoctorInfo(s.c, x.UserID)

	// return session token
	return &pb.DoctorData{
		Result: pb.DoctorData_Ok,
		Name: d.Name,
		Email: d.Email,
	}, nil
}


// Get Patients
func (s *server) PGetPatients(ctx context.Context, x *pb.UserID) (*pb.PatientsResponse, error) {
	log.Printf("GetPatients:")

	ds := myFire.GetPatientsOfDoctor(s.c, x.UserID)

	// return session token
	return &pb.PatientsResponse{
		Result: pb.PatientsResponse_Ok,
		Patients: ds,
	}, nil
}
// Get Test History
func (s *server) GetTestHistory(ctx context.Context, x *pb.UserID) (*pb.TestHistoryResponse, error) {
	log.Printf("GetTestHistory:")

	ts := myFire.GetTestHistory(s.c, x.UserID)

	// return session token
	return &pb.TestHistoryResponse{
		Result: pb.TestHistoryResponse_Ok,
		Patients: ts,
	}, nil
}

// Send Lifestyle Questionares
func (s *server) SendLifestyle(ctx context.Context, x *pb.LifestyleRequest) (*pb.LifestyleResponse, error) {
	log.Printf("SendLifestyle:")

	myFire.AddLifestyleTest(s.c, x.UserID, x.Data)

	// return session token
	return &pb.LifestyleResponse{
		Result: pb.LifestyleResponse_Ok,
	}, nil
}

// Send Patient Dementia
func (s *server) SendPatientDementia(ctx context.Context, x *pb.DementiaRequest) (*pb.DementiaResponse, error) {
	log.Printf("SendPatientDementia:")

	myFire.SetPatientDementica(s.c, x.UserID, x.Dementia)

	// return session token
	return &pb.LifestyleResponse{
		Result: pb.SendPatientDementia_Ok,
	}, nil
}

// GetNews
func (s *server) GetNews(ctx context.Context, x *pb.NewsRequest) (*pb.NewsResponse, error) {
	log.Printf("GetNews: __NotImplemented__")

	// todo: check type

	return &pb.NewsResponse{
		Content: "News Not Implemented Yet",
	}, nil
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
