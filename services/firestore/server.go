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
	firebase "firebase.google.com/go"
	firestore "cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)


///////////////////////////////////////////////////////////////
/// Temp
type UserRecord struct {
	Username string
	Password string
	RiskFactor int32
}
// TODO, test date

///////////////////////////////////////////////////////////////
/// Firestore

func firebaseInit() *firestore.Client {
	ctx := context.Background()
	creds := option.WithCredentialsFile("./firebase.json")
	app, err := firebase.NewApp(ctx,nil,creds)
	if err != nil { log.Fatalf("Firebase: failed to create app:\n%v",err)}
	// was var err2 error;
	client, err := app.Firestore(ctx)
	if err != nil { log.Fatalf("Firebase: failed to access store:\n%v",err)}
	return client
}

func DeleteCollection(c *firestore.Client, colName string){
	ctx := context.Background()
	iter := c.Collection(colName).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done { break }
		if err != nil { log.Fatalf("failed to iterate:\n%v",err)}
		doc.Ref.Delete(ctx)
	}
}

func BURN_IT_ALL_DOWN(c *firestore.Client) {
	DeleteCollection(c,"Users")
	DeleteCollection(c,"TestResults")

	docId := RegisterDoctor(c,"Eoin","Eoin@NeuroMind.com","12345")
	patId := RegisterPatient(c,"Conor","12345",docId)
	_ = RegisterPatient(c,"Conor2","12345",docId)
	v, docId2, logType := Login(c,"Eoin","12345")
	info, e := GetDoctorInfo(c,docId2)
	info2, e2 := GetPatientInfo(c,patId)
	ps := GetPatientsOfDoctor(c,docId)

	AddLifestyleTest(c, patId, "Diabetic:true,AlcoholLevel:0.084973629, HeartRate:98, BloodOxygenLevel:96.23074296, BodyTemperature:36.22485168, Weight:57.56397754, MRI_Delay:36.42102798, Presecription:None, DosageMg:0, Age:60, EducationLevel:Primary School, DominantHand:Left, Gender:Female, FamilyHistory:false, SmokingStatus:Current Smoker, APOE_e19:false, PhysicalActivity:Sedentary, DepressionStatus:false, MedicationHistory:false, NutritionDiet:Low-Carb Diet, SleepQuality:Poor, ChronicHealthConditionsDiabetes" )
	ts := GetRiskScoreHistory(c, patId)
	SetPatientDementica(c,patId, "Positive")
	info3,e3 := GetPatientInfo(c, patId)


	log.Printf("%v",docId)
	log.Printf("%v",patId)
	log.Printf("%v-%v-%v",v,docId2,logType)
	log.Printf("%v-%v",info,e)
	log.Printf("%v-%v",info2,e2)
	log.Printf("list: %v",ps)
	log.Printf("tests: %v",ts)
	log.Printf("patient2: %v-%v",info3,e3)
}

// DocumentRef, WriteResult, err := client.Collection("NAME").Add(ctx, map[string]interface{}{
// 	"FIELD":var,
// 	"FIELD2":var2
// })



// iter := client.Collection("NAME").Documents(ctx)
// for {
// 	// iterate
// 	doc, err := iter.Next() // returns document snapshot
// 	if err == iterator.Done { break }
// 	if err != nil { log.Fatalf("failed to iterate:\n%v",err)}

// // Get Field
// post := entity.Post { ID:doc.Data()["ID"].(int), ... } // throw error if not given type (["field"] might return nil)
// posts = append(posts,post)

// // Auto Get Struct
// var docData UserRecord
// if err := doc.DataTo(&docData); err != nil {
//	log.Fatalf("err2") }



type Patient struct {
	UserID string
	Name string
	HasDementia string
	HasDoctor bool
	DoctorID string
	HasRiskScore bool
	RiskScore float64 // change to string of none
}

type Doctor struct {
	UserID string
	Name string
	Email string
}
type TestResult struct {
	Date string
	TestID string
	RiskScore string
}

// TestResult(Date,testID,RiskScore)
func EmptyDoctor() Doctor { return Doctor{UserID:"",Name:"",Email:""} }
func EmptyPatient() Patient { return Patient{UserID:"",Name:"", HasDementia:"Unknown", HasDoctor:false, DoctorID:"", HasRiskScore:false, RiskScore:0.0}}



// Returns DocumentID
func RegisterDoctor(c *firestore.Client, name string, email string, password string) string{
	ctx := context.Background()

	docRef, _, err := c.Collection("Users").Add(ctx, map[string]interface{}{
		"Name":name,
		"Email":email,
		"Password":password,
		"Type":"Doctor",
	})
	if err != nil { log.Fatalf("RegisterDoc error\n%v",err)}
	log.Printf("Registered Doctor: %s-%s-%s, %v\n\n",name,email,password, docRef)
	return docRef.ID
}

func RegisterPatient(c *firestore.Client, name string, password string, doctorRef string) string{
	ctx := context.Background()

	docRef, _, err := c.Collection("Users").Add(ctx, map[string]interface{}{
		"Name":name,
		"Password":password,
		"HasDementia":"Unknown",
		"DoctorID":doctorRef,
		"RiskScore":0.5,
		"Type":"Patient",
	})
	if err != nil { log.Fatalf("RegisterPatient error\n%v",err)}
	log.Printf("Registered Patient: %s-%s, %v\n\n",name,password, docRef)
	return docRef.ID
}



// Returns valid,UserID,type
func Login(c *firestore.Client, name string, password string) (bool,string,string){
	ctx := context.Background()

	valid := true;
	userId := "";
	userType := "";

	iter := c.Collection("Users").Documents(ctx)
	for {
		// iterate
		doc, err := iter.Next()
		if err == iterator.Done { break }
		if err != nil { log.Fatalf("failed to iterate:\n%v",err)}

		if doc.Data()["Name"].(string) != name { continue; }
		if doc.Data()["Password"].(string) != password { valid = false; }
		userType = doc.Data()["Type"].(string)
		userId = doc.Ref.ID
	}
	return valid,userId,userType
}

// func Logout not implemented

func GetDoctorInfo(c *firestore.Client, ref string) (Doctor,bool){
	ctx := context.Background()
	doc, err := c.Collection("Users").Doc(ref).Get(ctx)
	if err != nil { log.Fatalf("GetDoctorInfo error(%s)\n%v",ref,err)}
	d := doc.Data()
	if d["Type"] != "Doctor" { return EmptyDoctor(), true }
	return Doctor{ UserID: ref, Name:d["Name"].(string), Email:d["Email"].(string) }, false
}
func GetPatientInfo(c *firestore.Client, ref string) (Patient, bool) {
	ctx := context.Background()
	doc, err := c.Collection("Users").Doc(ref).Get(ctx)
	if err != nil { log.Fatalf("GetPatietnInfo error\n%c",err)}
	d := doc.Data()
	if d["Type"] != "Patient" { return EmptyPatient(), true }
	return Patient{ UserID: ref, Name:d["Name"].(string), HasDementia:d["HasDementia"].(string), HasDoctor:true, DoctorID:d["DoctorID"].(string), HasRiskScore:true, RiskScore: d["RiskScore"].(float64) }, false
}

func GetPatientsOfDoctor(c *firestore.Client, docId string) []Patient{
	ctx := context.Background()

	var out []Patient;
	iter := c.Collection("Users").Documents(ctx)
	for {
		// iterate
		doc, err := iter.Next()
		if err == iterator.Done { break }
		if err != nil { log.Fatalf("failed to iterate:\n%v",err)}
		d := doc.Data()

		if d["Type"].(string) != "Patient" { continue; }
		if d["DoctorID"].(string) == docId {
			p := Patient{ UserID: doc.Ref.ID, Name:d["Name"].(string), HasDementia:d["HasDementia"].(string), HasDoctor:true, DoctorID:d["DoctorID"].(string), HasRiskScore:true, RiskScore: d["RiskScore"].(float64) }
			out = append(out,p)
		}
	}
	return out
}

func AddLifestyleTest(c *firestore.Client, patId string, lifestyle string) {
	ctx := context.Background()

	_, _, err := c.Collection("TestResults").Add(ctx, map[string]interface{}{
		"UserID":patId,
		"Date":"__NOT__IMPLEMENTED__",
		"RiskScore":"Calculating",
		"Data":lifestyle,
	})
	if err != nil { log.Fatalf("RegisterPatient error\n%v",err)}

}
func GetRiskScoreHistory(c *firestore.Client, patId string) []TestResult {
	ctx := context.Background()

	var out []TestResult;
	iter := c.Collection("TestResults").Documents(ctx)
	for {
		// iterate
		doc, err := iter.Next()
		if err == iterator.Done { break }
		if err != nil { log.Fatalf("failed to iterate:\n%v",err)}
		d := doc.Data()

		if d["UserID"].(string) != patId { continue; }
		tr := TestResult{ Date:d["Date"].(string), TestID:doc.Ref.ID, RiskScore:d["RiskScore"].(string) }
		out = append(out,tr)
	}
	return out
}

func SetPatientDementica(c *firestore.Client, patId string, dementica string) {
	ctx := context.Background()
	_, err := c.Collection("Users").Doc(patId).Update(ctx, []firestore.Update{{
		Path: "HasDementia", Value:dementica},})
	if err != nil { log.Printf("Error %v", err) }
}





// Patient(UserID,Name,HasDementia,DoctorID?,RiskScore?)
// Doctor(UserID,Name,Email)
// TestResult(Date,testID,RiskScore)

// RegisterDoctor(name,email,password) -> uid
// Login(name,password) -> x,uid
// Logout(x)
// GetDoctorInfo(uid) -> (UserID,Name,Email)
// GetPatientInfo(uid) -> (UserID,Name,HasDementia,DoctorID?,RiskScore?)
// GetPatientsOfDoctor(uid) -> list(patiend)
// GetRiskScoreHistory(uid) -> list(TestResult)
// TOOD: GetTestData(testID) -> xxx
// SetPatientDementica(uid, [Unknown,Positive,Negative]) -> None
// GetNews([Doctor,Patient]) -> str



///////////////////////////////////////////////////////////////
/// Old

// Firebase



// Auth/Tokens
type Session_Tokens_Type struct {
	data [65535]string
	mu sync.RWMutex
	idx int64
	// todo free list, more info stored about token not just username
}
var Session_Tokens = Session_Tokens_Type{ idx: 0}


func ValidateLogin(username string, password string) bool {
	return true }



///////////////////////////////////////////////////////////////
/// GRPC
type server struct{
	pb.UnimplementedFirestoreServer
	client *firestore.Client
}



// Login (UserLogin -> SessionToken)
func (s *server) Login(ctx context.Context, x *pb.UserLogin) (*pb.SessionToken, error) {
	log.Printf("login: %s, %s\n\n",x.UserName, x.PlaintextPassword)

	// Mutex Write Lock
	Session_Tokens.mu.Lock()
	defer Session_Tokens.mu.Unlock()

	// Validate Login and assign token
	if ValidateLogin(x.UserName, x.PlaintextPassword) {
		Session_Tokens.data[Session_Tokens.idx] = x.UserName
	} else { Session_Tokens.data[Session_Tokens.idx] = "__invalid__"}

	returnVal := Session_Tokens.idx
	Session_Tokens.idx = Session_Tokens.idx +1

	// return session token
	return &pb.SessionToken{
		Temp: returnVal,
	}, nil
}



// GetDetails (UserRequest -> UserDetails)
func (s *server) GetDetails(ctx context.Context, x *pb.UserRequest) (*pb.UserDetails, error) {
	log.Printf("GetDetails: %s, %s", x.SessionToken, x.UserId)

	// Mutex Read Lock
	Session_Tokens.mu.RLock()
	defer Session_Tokens.mu.RUnlock()

	idx := x.SessionToken

	// check if user can access required data
	log.Printf("Idx: %d, UserId: %s, Session: %s\n\n",idx,x.UserId, Session_Tokens.data[idx])
	if Session_Tokens.data[idx] == x.UserId {
		return &pb.UserDetails{
			Details:"some details" }, nil
	}
	// unauthorized access response
	return &pb.UserDetails{
		Details:"invalid" }, nil
}



// GetRisk (SessionToken -> RiskScore)
func (s *server) GetRisk(ctx context.Context, x *pb.SessionToken) (*pb.RiskScore, error) {

	// Mutex Read Lock
	Session_Tokens.mu.RLock()
	username := Session_Tokens.data[x.Temp] // todo, validate valid session Token (not out of bounds etc)
	Session_Tokens.mu.RUnlock()

	log.Printf("GetRisk: Session: %d username: %s", x.Temp, username)

	// find
	iter := s.client.Collection("users").Documents(context.Background())
	for { // todo: probably a way to do this on server
		// iterate
		doc, err := iter.Next()
		if err == iterator.Done { break }
		if err != nil { log.Fatalf("failed to iterate:\n%v",err)}

		// get data of record
		var docData UserRecord
		if err := doc.DataTo(&docData); err != nil {
			log.Fatalf("err2") }

		// check if target user
		if docData.Username == username {
			log.Printf("%d",docData.RiskFactor)
			return &pb.RiskScore{ Score: docData.RiskFactor, }, nil
		}
	}

	// Dummy Response
	return &pb.RiskScore{ Score: 0, }, nil
}



func (s *server) ProcessLifestyle(x string) (string,bool) {
	// return "0" // probably better to not reconnect each time idk?

	print("firebase attempt")
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
 


// } // todo, grpc into vertexAI

// SendLifestyle (SessionToken -> RiskScore)
func (s *server) SendLifestyle(ctx context.Context, x *pb.LifestyleRequest) (*pb.LifestyleResponse, error) {

	// Mutex Read Lock
	// Session_Tokens.mu.RLock()
	// username := Session_Tokens.data[x.Temp] // todo, validate valid session Token (not out of bounds etc)
	// Session_Tokens.mu.RUnlock()

	log.Printf("SendLifestyle:'%s'", x.Message)


	FBctx := context.Background()
	calc_risk, ok := s.ProcessLifestyle(x.Message) // vertexAI
	if ok == false { calc_risk = "Error Calculating"; }
	log.Printf("calc_risk: %s\n",calc_risk)
	// TODO: log errors that occur in database, or some log file

	// test firebase add
	// DocumentRef, WriteResult,error
	_, _, err2 := s.client.Collection("patientData").Add(FBctx, map[string]interface{}{
		"data":x.Message,
		"calculated_risk":calc_risk,
	})
	if err2 != nil { log.Fatalf("Failed adding\n%v", err2)}


	// find
	// iter := client.Collection("patientData").Documents(context.Background())
	// for { // todo: probably a way to do this on server
	// 	// iterate
	// 	doc, err := iter.Next()
	// 	if err == iterator.Done { break }
	// 	if err != nil { log.Fatalf("failed to iterate:\n%v",err)}

	// 	// get data of record
	// 	var docData UserRecord
	// 	if err := doc.DataTo(&docData); err != nil {
	// 		log.Fatalf("err2") }

	// 	// check if target user
	// 	if docData.Username == username {
	// 		log.Printf("%d",docData.RiskScore)
	// 		return &pb.RiskScore{ Score: docData.RiskScore, }, nil
	// 	}
	// }

	// Dummy Response
	return &pb.LifestyleResponse{ Success: true, }, nil
}







///////////////////////////////////////////////////////////////
/// global firebase client, initialized at startup






///////////////////////////////////////////////////////////////
/// Main

func main() {
	// grpc connection
	lis, err := net.Listen("tcp", ":9000");
	if err != nil { log.Fatalf("GRPC: failed to listen:\n%v", err) }

	// serv GRPC
	serverData := server{client:firebaseInit()}
	defer serverData.client.Close()
	BURN_IT_ALL_DOWN(serverData.client)


	grpcServer := grpc.NewServer()
	pb.RegisterFirestoreServer(grpcServer, &serverData)
	log.Printf("Ready!! >:0")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("GRPC: Failed to serve:\n%v",err) }

}
