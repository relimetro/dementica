package myFire

import(
	"log" // for loggin
	"google.golang.org/api/iterator"

	// Grpc
	"golang.org/x/net/context"

	firebase "firebase.google.com/go"
	// firestore "cloud.google.com/go/firestore"
	"google.golang.org/api/option"

	// firebase
	firestore "cloud.google.com/go/firestore"
)

///////////////////////////////////////////////////////////////
/// Firestore

func FirebaseInit() *firestore.Client {
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
	log.Printf("BURN_IT_ALL_DOWN")

	DeleteCollection(c,"Users")
	DeleteCollection(c,"TestResults")

	docId := RegisterDoctor(c,"Eoin","Eoin@NeuroMind.com","12345")
	patId := RegisterPatient(c,"Conor","12345",docId)
	_ = RegisterPatient(c,"Conor2","12345",docId)
	info, e := GetDoctorInfo(c,docId)
	info2, e2 := GetPatientInfo(c,patId)
	ps := GetPatientsOfDoctor(c,docId)

	AddLifestyleTest(c, patId, "Diabetic:true,AlcoholLevel:0.084973629, HeartRate:98, BloodOxygenLevel:96.23074296, BodyTemperature:36.22485168, Weight:57.56397754, MRI_Delay:36.42102798, Presecription:None, DosageMg:0, Age:60, EducationLevel:Primary School, DominantHand:Left, Gender:Female, FamilyHistory:false, SmokingStatus:Current Smoker, APOE_e19:false, PhysicalActivity:Sedentary, DepressionStatus:false, MedicationHistory:false, NutritionDiet:Low-Carb Diet, SleepQuality:Poor, ChronicHealthConditionsDiabetes" )
	// ts := GetRiskScoreHistory(c, patId)
	ts := GetTestHistory(c, patId)

	log.Printf("%v",docId)
	log.Printf("%v",patId)
	// log.Printf("%v-%v-%v",v,docId2,logType)
	log.Printf("%v-%v",info,e)
	log.Printf("%v-%v",info2,e2)
	log.Printf("list: %v",ps)
	log.Printf("tests: %v",ts)
	// log.Printf("patient2: %v-%v",info3,e3)
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
	HasDementia string // Unknown, Positive, Negative
	DoctorID string
	HasRiskScore bool
	RiskScore string // "" if empty
}

type Doctor struct {
	UserID string
	Name string
	Email string
}
type TestResult struct {
	Date string
	// TestID string
	RiskScore string // "Calculating" if calculating
}

// TestResult(Date,testID,RiskScore)
func EmptyDoctor() Doctor { return Doctor{UserID:"",Name:"",Email:""} }
func EmptyPatient() Patient { return Patient{UserID:"",Name:"", HasDementia:"Unknown", DoctorID:"", HasRiskScore:false, RiskScore:"0.0"}}



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
		"RiskScore":"0.5",
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
	return Patient{ UserID: ref, Name:d["Name"].(string), HasDementia:d["HasDementia"].(string), DoctorID:d["DoctorID"].(string), HasRiskScore:true, RiskScore:d["RiskScore"].(string) }, false
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
			p := Patient{ UserID: doc.Ref.ID, Name:d["Name"].(string), HasDementia:d["HasDementia"].(string), DoctorID:d["DoctorID"].(string), HasRiskScore:true, RiskScore:d["RiskScore"].(string) }
			out = append(out,p)
		}
	}
	return out
}

// returns document ID
func AddLifestyleTest(c *firestore.Client, patId string, lifestyle string) string {
	ctx := context.Background()

	docRef, _, err := c.Collection("TestResults").Add(ctx, map[string]interface{}{
		"UserID":patId,
		"Date":"__NOT__IMPLEMENTED__",
		"RiskScore":"Calculating",
		"Data":lifestyle,
	})
	if err != nil { log.Fatalf("RegisterPatient error\n%v",err)}
	return docRef.ID

}
func GetTestHistory(c *firestore.Client, patId string) []TestResult {
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
		tr := TestResult{ Date:d["Date"].(string), RiskScore:d["RiskScore"].(string) }
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

func SetTestRiskScore(c *firestore.Client, testId string, riskScore string) {
	ctx := context.Background()
	_, err := c.Collection("TestResults").Doc(testId).Update(ctx, []firestore.Update{{
		Path: "RiskScore", Value:riskScore},})
	if err != nil { log.Printf("Error %v", err) }
}



func GetNews(c *firestore.Client, matchStr string) string{
	ctx := context.Background()

	iter := c.Collection("News").Documents(ctx)
	for {
		// iterate
		doc, err := iter.Next()
		if err == iterator.Done { break }
		if err != nil { log.Fatalf("failed to iterate:\n%v",err)}
		d := doc.Data()

		if d["Type"].(string) == matchStr {
			return d["Content"].(string) }
	}
	return "Sorry, no news found for type '"+matchStr+"'."
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
