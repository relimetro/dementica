package main

import(
	"log" // for loggin

	// Grpc
	"net"
	pb "example/proto_example/protoOut"
	aiProompt "example/proto_example/protoAI"
	UserService "example/proto_example/UserService"
	"google.golang.org/grpc"
	"golang.org/x/net/context"

	// firebase
	firestore "cloud.google.com/go/firestore"
	"example/proto_example/myFire"
)



///////////////////////////////////////////////////////////////
/// Functionality

func KerasCall(x string) (string,bool) {

	var conn *grpc.ClientConn
	// conn, err := grpc.Dial("vertexai:50052", grpc.WithInsecure())
	conn, err := grpc.Dial("keras:50053", grpc.WithInsecure())
	if err != nil { log.Printf("[ERROR] GRPC: cound not connect keras at 50053: \n%s",err); return "",false; }
	defer conn.Close()
	c := aiProompt.NewAiProomptClient(conn)

	// txt := "Diabetic:true,AlcoholLevel:0.084973629, HeartRate:98, BloodOxygenLevel:96.23074296, BodyTemperature:36.22485168, Weight:57.56397754, MRI_Delay:36.42102798, Presecription:None, DosageMg:0, Age:60, EducationLevel:Primary School, DominantHand:Left, Gender:Female, FamilyHistory:false, SmokingStatus:Current Smoker, APOE_e19:false, PhysicalActivity:Sedentary, DepressionStatus:false, MedicationHistory:false, NutritionDiet:Low-Carb Diet, SleepQuality:Poor, ChronicHealthConditionsDiabetes"
	log.Printf("calling keras (%s)",x)
	message := aiProompt.ProomptMsg { Message: x}
	resp, err := c.HealtcareProompt(context.Background(), &message)
	if err != nil { log.Printf("[ERROR] FTproompt, <%s>, <%d>",err,resp); return "",false; }
	log.Printf("Response FTproompt: %s",resp)
	return resp.Message,true }



func GoPatientToProtoPatient(p myFire.Patient) pb.PatientData {
	hasDem := pb.PatientData_Unknown;
	if ( p.HasDementia == "Positive" ) {
		hasDem = pb.PatientData_Positive;
	} else if ( p.HasDementia == "Negative") { hasDem = pb.PatientData_Negative;
	} else if ( p.HasDementia == "Unknown") { hasDem = pb.PatientData_Unknown;
	} else { log.Printf("ERROR: string(%s), could not be converted to dementia type",p.HasDementia) }

	return pb.PatientData{
		Result: pb.PatientData_Ok,
		Name: p.Name,
		HasDementia: hasDem,
		DoctorID: p.DoctorID,
		RiskScore: p.RiskScore,
	}

}


///////////////////////////////////////////////////////////////
/// UserService
func (s *server) VerifyIdToken(txt string) (string,bool) {
	var conn *grpc.ClientConn
	// user_service
	conn, err := grpc.Dial("localhost:50061", grpc.WithInsecure())
	if err != nil { log.Printf("[ERROR] GRPC: cound not connect user_service at 50061: \n%s",err); return "",false; }
	defer conn.Close()
	c := UserService.NewUserServiceClient(conn)

	message := UserService.VerifyTokenRequest { IdToken: txt} // note GO: removed underscore and capitalizes id_token -> Id_Token
	resp, err := c.VerifyTokenRemote(context.Background(), &message)
	if err != nil { log.Printf("[ERROR] firestore verify id_token, <%s>, <%d>",err,resp); return "",false; }
	log.Printf("Response id_token verify: %v-%s",resp.Res,resp.Uid)
	return resp.Uid,true }

// returns id_token
func (s *server) UserService_Register(email string, password string) (string,bool) {
	var conn *grpc.ClientConn
	// user_service
	conn, err := grpc.Dial("localhost:50061", grpc.WithInsecure())
	if err != nil { log.Printf("[ERROR] GRPC: cound not connect user_service at 50061: \n%s",err); return "",false; }
	defer conn.Close()
	c := UserService.NewUserServiceClient(conn)

	message := UserService.SignUpRequest { Email: email, Password:password}
	resp, err := c.SignUp(context.Background(), &message)
	if err != nil { log.Printf("[ERROR] firestore user_service signup, <%s>, <%d>",err,resp); return "",false; }
	log.Printf("Response user_service signup: %s",resp.Message)
	return "AHHASH",true }

func (s *server) UserService_Login(email string, password string) bool {
	var conn *grpc.ClientConn
	// user_service
	conn, err := grpc.Dial("localhost:50061", grpc.WithInsecure())
	if err != nil { log.Printf("[ERROR] GRPC: cound not connect user_service at 50061: \n%s",err); return false; }
	defer conn.Close()
	c := UserService.NewUserServiceClient(conn)

	message := UserService.LoginRequest { Email: email, Password: password }
	resp, err := c.Login(context.Background(), &message)
	if err != nil { log.Printf("[ERROR] firestore user_service login, <%s>, <%d>",err,resp); return false; }
	log.Printf("Response user_service signup: %s",resp.Message)
	return true }





///////////////////////////////////////////////////////////////
/// GRPC
type server struct{
	pb.UnimplementedFirestoreServer
	c *firestore.Client
}



// Register
func (s *server) Register(ctx context.Context, x *pb.UserRegister) (*pb.RegisterResult, error) {
	log.Printf("register:") //%s, %s\n\n",x.UserName, x.PlaintextPassword)

	if x.UserType == pb.UserRegister_Patient {
		myFire.RegisterPatient(s.c, x.Name, x.Password, "") // todo ability to assign to doctor
	} else {
		myFire.RegisterDoctor(s.c, x.Name, x.Password, x.RegisterWith) // todo check for correct email etc
	}


	// return session token
	return &pb.RegisterResult{
		Result: pb.RegisterResult_Ok,
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
		Result: pb.LoginResult_Ok,
	}, nil
}





// PatientInfo
func (s *server) PatientInfo(ctx context.Context, x *pb.UserID) (*pb.PatientData, error) {
	log.Printf("patientInfo:")

	p, _ := myFire.GetPatientInfo(s.c, x.UserID)

	p2 := GoPatientToProtoPatient(p)

	// return session token
	return &p2, nil
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
func (s *server) GetPatients(ctx context.Context, x *pb.UserID) (*pb.PatientsResponse, error) {
	log.Printf("GetPatients:")

	ds := myFire.GetPatientsOfDoctor(s.c, x.UserID)

	var ds2 []*pb.PatientData

	for _,x := range(ds) {
		temp := GoPatientToProtoPatient(x)
		ds2 = append(ds2,&temp)
	}

	// return session token
	return &pb.PatientsResponse{
		Result: pb.PatientsResponse_Ok,
		Patients: ds2,
	}, nil
}
// Get Test History
func (s *server) GetTestHistory(ctx context.Context, x *pb.UserID) (*pb.TestHistoryResponse, error) {
	log.Printf("GetTestHistory:")

	ts := myFire.GetTestHistory(s.c, x.UserID)

	var ts2 []*pb.TestData
	for _,x := range(ts){
		temp := pb.TestData{
			Date: x.Date,
			RiskScore: x.RiskScore,
		}
		ts2 = append(ts2,&temp)

	}

	// return session token
	return &pb.TestHistoryResponse{
		Result: pb.TestHistoryResponse_Ok,
		Tests: ts2,
	}, nil
}

// Send Lifestyle Questionares
func (s *server) SendLifestyle(ctx context.Context, x *pb.LifestyleRequest) (*pb.LifestyleResponse, error) {
	log.Printf("SendLifestyle: '%s'",x.Data)

	myFire.AddLifestyleTest(s.c, x.UserID, x.Data)
	out,succ := KerasCall(x.Data)

	// todo go func
	log.Printf("%v -- %v",out,succ)
	// and update riskScore in background

	// return session token
	return &pb.LifestyleResponse{
		Result: pb.LifestyleResponse_Ok,
	}, nil
}

// Send Patient Dementia
func (s *server) SendPatientDementia(ctx context.Context, x *pb.DementiaRequest) (*pb.DementiaResponse, error) {
	log.Printf("SendPatientDementia:")

	demStr := "Unknown"
	if x.Dementia == pb.DementiaRequest_Unknown { demStr = "Unknown" 
	} else if x.Dementia == pb.DementiaRequest_Positive { demStr = "Positive" 
	} else if x.Dementia == pb.DementiaRequest_Negative { demStr = "Negative" 
	} else { log.Printf("ERROR: failed to convert DementiaRequest.dementia (%v) into string",x.Dementia)}

	myFire.SetPatientDementica(s.c, x.UserID, demStr)

	// return session token
	return &pb.DementiaResponse{
		Result: pb.DementiaResponse_Ok,
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
	serverData := server{c:myFire.FirebaseInit()}
	defer serverData.c.Close()

	// Reset firestore (note not integrated with danis stuff (linking, auth))
	// myFire.BURN_IT_ALL_DOWN(serverData.c)


	// start server
	grpcServer := grpc.NewServer()
	pb.RegisterFirestoreServer(grpcServer, &serverData)
	log.Printf("Ready!! >:0")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("GRPC: Failed to serve:\n%v",err) }
}
