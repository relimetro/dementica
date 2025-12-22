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
	resp, err := c.Proompt(context.Background(), &message)
	if err != nil { log.Printf("[ERROR] FTproompt, <%s>, <%d>",err,resp); return "",false; }
	log.Printf("Response FTproompt: %s",resp)
	return resp.Message,true }

func UpdateTestRiskScore(c *firestore.Client, data string, testId string, patId string) (string,bool) { // risk score, success
	out,succ := KerasCall(data)
	log.Printf("UpdateTestRiskScore: %v -- %v",succ,out)
	if succ {
		myFire.SetTestRiskScore(c, testId, out)
		myFire.SetPatientRisk(c,patId,out)
		return out,true
	} else {
		myFire.SetTestRiskScore(c, testId, "ServerError") 
		return "",false
	}
	// log.Printf("UpdateTestRiskScore: Done") // no longer async
}



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



func VertexCall(x string) (string,bool) {

	var conn *grpc.ClientConn
	// conn, err := grpc.Dial("vertexai:50052", grpc.WithInsecure())
	conn, err := grpc.Dial("vertexai:50052", grpc.WithInsecure())
	if err != nil { log.Printf("[ERROR] GRPC: cound not connect vertexai at 50052: \n%s",err); return "",false; }
	defer conn.Close()
	c := aiProompt.NewAiProomptClient(conn)

	log.Printf("calling vertexai (%s)",x)
	message := aiProompt.ProomptMsg { Message: x}
	resp, err := c.Proompt(context.Background(), &message)
	if err != nil { log.Printf("[ERROR] FTproompt, <%s>, <%d>",err,resp); return "",false; }
	log.Printf("Response FTproompt: %s",resp)
	return resp.Message,true }

func UpdateTranscriptRiskScore(c *firestore.Client, data string, testId string) (string,bool) { // riskScore,success
	out,succ := VertexCall(data)
	log.Printf("UpdateTranscrip tRiskScore: %v -- %v",succ,out)
	if succ {
		myFire.SetTestRiskScore(c, testId, out)
		myFire.SetPatientRisk(c,testId,out)
		return out,true
	} else {
		myFire.SetTestRiskScore(c, testId, "ServerError") 
		return "",false }
	}


///////////////////////////////////////////////////////////////
/// UserService
func (s *server) VerifyIdToken(idToken string) (string,bool) { // returns uid of given idToken
	var conn *grpc.ClientConn
	// user_service
	conn, err := grpc.Dial("user_service:50061", grpc.WithInsecure())
	if err != nil { log.Printf("[ERROR] GRPC: cound not connect user_service at 50061: \n%s",err); return "",false; }
	defer conn.Close()
	c := UserService.NewUserServiceClient(conn)

	message := UserService.VerifyTokenRequest { IdToken: idToken} // note GO: removed underscore and capitalizes id_token -> Id_Token
	resp, err := c.VerifyTokenRemote(context.Background(), &message)
	if err != nil { log.Printf("[ERROR] firestore verify id_token, <%s>, <%d>",err,resp); return "",false; }
	log.Printf("Response id_token verify: %v-%s",resp.Res,resp.Uid)
	return resp.Uid,true }


func (s *server) UserService_Register(email string, password string) (string,bool) { // returns uid and success
	var conn *grpc.ClientConn
	// user_service
	conn, err := grpc.Dial("user_service:50061", grpc.WithInsecure())
	if err != nil { log.Printf("[ERROR] GRPC: cound not connect user_service at 50061: \n%s",err); return "",false; }
	defer conn.Close()
	c := UserService.NewUserServiceClient(conn)

	message := UserService.SignUpRequest { Email: email, Password:password}
	resp, err := c.SignUp(context.Background(), &message)
	if err != nil { log.Printf("[ERROR] firestore user_service signup, <%s>, <%d>",err,resp); return "",false; }
	log.Printf("Response user_service signup: %s",resp)
	if resp.Message == "Signup failed" {
		return "", false; }
	return resp.Uid,true }


func (s *server) UserService_Login(email string, password string) (string,string,bool) { // returns id_token, uid, and success
	var conn *grpc.ClientConn
	// user_service
	conn, err := grpc.Dial("user_service:50061", grpc.WithInsecure())
	if err != nil { log.Printf("[ERROR] GRPC: cound not connect user_service at 50061: \n%s",err); return "","",false; }
	defer conn.Close()
	c := UserService.NewUserServiceClient(conn)

	message := UserService.LoginRequest { Email: email, Password: password }
	resp, err := c.Login(context.Background(), &message)
	if err != nil { log.Printf("[ERROR] firestore user_service login, <%s>, <%d>",err,resp); return "","",false; }
	log.Printf("Response user_service signup: [%s] <%s>",resp.Message,resp.IdToken)
	return resp.IdToken, resp.Uid, true }





///////////////////////////////////////////////////////////////
/// GRPC
type server struct{
	pb.UnimplementedFirestoreServer
	c *firestore.Client
}



// Register
func (s *server) Register(ctx context.Context, x *pb.UserRegister) (*pb.RegisterResult, error) {
	log.Printf("register: %s %s",x.Name,x.RegisterWith,x.Password,) //%s, %s\n\n",x.UserName, x.PlaintextPassword)

	// only allow email
	if x.RegType != pb.UserRegister_Email {
		log.Printf("ERROR, only email registration is supported")
		return &pb.RegisterResult{
			Result: pb.RegisterResult_Failed,
		}, nil }

	uid, succ := s.UserService_Register(x.RegisterWith,x.Password)
	if succ {
		if x.UserType == pb.UserRegister_Patient {
			myFire.RegisterPatient(s.c, uid, x.Name, x.RegisterWith)
		} else {
			myFire.RegisterDoctor(s.c, uid, x.Name, x.RegisterWith); }

		// return Failed
		return &pb.RegisterResult{
			Result: pb.RegisterResult_Ok,
		}, nil
	} else {
		// return Failed
		return &pb.RegisterResult{
			Result: pb.RegisterResult_Taken,
		}, nil
	}


}

// Login (UserLogin -> LoginResult)
func (s *server) Login(ctx context.Context, x *pb.UserLogin) (*pb.LoginResult, error) {
	log.Printf("login: %s, %s\n\n",x.Email, x.Password)
	// x.UserType

	idToken, uid, succ := s.UserService_Login(x.Email,x.Password)
	if succ {
		return &pb.LoginResult{
			IdToken: idToken,
			UserID: uid,
			Result: pb.LoginResult_Ok,
		}, nil
	} else {
		return &pb.LoginResult{
			IdToken: "",
			UserID: "",
			Result: pb.LoginResult_UserPass,
		}, nil
	}
}





// PatientInfo
func (s *server) PatientInfo(ctx context.Context, x *pb.UserID) (*pb.PatientData, error) {
	log.Printf("patientInfo: %s",x.UserID)

	p, _ := myFire.GetPatientInfo(s.c, x.UserID)

	p2 := GoPatientToProtoPatient(p)

	// return session token
	return &p2, nil
}
// Get Risk
func (s *server) GetRisk(ctx context.Context, x *pb.UserID) (*pb.RiskResponse, error) {
	log.Printf("GetRisk: %s", x.UserID)

	p, _ := myFire.GetPatientInfo(s.c, x.UserID)

	score := p.RiskScore
	if (score == "ServerError") { score = "Calculating" } // server error occurs when keras offline, convert to generic "Calculating" for users

	// return session token
	return &pb.RiskResponse{
		Result: pb.RiskResponse_Ok,
		RiskScore: p.RiskScore,
	}, nil
}


// Doctor Info
func (s *server) DoctorInfo(ctx context.Context, x *pb.UserID) (*pb.DoctorData, error) {
	log.Printf("DoctorInfo: %s", x.UserID)

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
	log.Printf("GetPatients: %s", x.UserID)

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
	log.Printf("GetTestHistory: %s", x.UserID)

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

	// upload test
	testId := myFire.AddLifestyleTest(s.c, x.UserID, x.Data, x.DateTime)

	// update risk score in background
	riskScore,succ := UpdateTestRiskScore(s.c, x.Data, testId, x.UserID)


	if succ {
		return &pb.LifestyleResponse{
			Result: pb.LifestyleResponse_Ok,
			RiskScore: riskScore,
		}, nil
	} else {
		return &pb.LifestyleResponse{
			Result: pb.LifestyleResponse_Error,
			RiskScore: "ServerError",
		}, nil
	}

}
// Send Transcript
func (s *server) SendTranscript(ctx context.Context, x *pb.LifestyleRequest) (*pb.LifestyleResponse, error) {
	log.Printf("SendTranscript: '%s'",x.Data)

	// upload test
	testId := myFire.AddTranscriptTest(s.c, x.UserID, x.Data, x.DateTime)

	// update risk score
	riskScore,succ := UpdateTranscriptRiskScore(s.c, x.Data, testId)
	if succ {
		return &pb.LifestyleResponse{
			Result: pb.LifestyleResponse_Ok,
			RiskScore: riskScore,
	}, nil
	} else {
		return &pb.LifestyleResponse{
			Result: pb.LifestyleResponse_Error,
			RiskScore: "ServerError",
		}, nil
	}
}

// Send Transcript
func (s *server) SendMinimental(ctx context.Context, x *pb.LifestyleRequest) (*pb.LifestyleResponse, error) {
	log.Printf("SendMinimental: '%s'",x.Data)

	// upload test
	myFire.AddMinimentalTest(s.c, x.UserID, x.Data, x.DateTime)

	return &pb.LifestyleResponse{
		Result: pb.LifestyleResponse_Ok,
		RiskScore: x.Data, }, nil
}

// Send Patient Dementia
func (s *server) SendPatientDementia(ctx context.Context, x *pb.DementiaRequest) (*pb.DementiaResponse, error) {
	log.Printf("SendPatientDementia: %s %s",x.UserID, x.Dementia)

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
	log.Printf("GetNews: %s",x.Type)

	var typeStr string
	if (x.Type == pb.NewsRequest_Patient) {
		typeStr = "Patient"
	} else if (x.Type == pb.NewsRequest_Doctor) { typeStr = "Doctor"
	} else {
		log.Printf("GetNews Invalid UserType\n%v",x)
		return &pb.NewsResponse{ Content: "Internal Server Error firestore-GetNews", }, nil
	}

	return &pb.NewsResponse{
		Content: myFire.GetNews(s.c,typeStr),
	}, nil
}

func (s *server) SetNews(ctx context.Context, x *pb.NewsSet) (*pb.NewsResponse, error) {
	log.Printf("SetNews: %s",x.Type)

	var typeStr string
	if (x.Type == pb.NewsSet_Patient) {
		typeStr = "Patient"
	} else if (x.Type == pb.NewsSet_Doctor) { typeStr = "Doctor"
	} else {
		log.Printf("GetNews Invalid UserType\n%v",x)
		return &pb.NewsResponse{ Content: "Internal Server Error firestore-GetNews", }, nil
	}

	myFire.SetNews(s.c,typeStr,x.Content )

	return &pb.NewsResponse{
		Content: "Ok",
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
