
txt:str = "Diabetic:true,AlcoholLevel:0.084973629, HeartRate:98, BloodOxygenLevel:96.23074296, BodyTemperature:36.22485168, Weight:57.56397754, MRI_Delay:36.42102798, Presecription:None, DosageMg:0, Age:60, EducationLevel:Primary School, DominantHand:Left, Gender:Female, FamilyHistory:false, SmokingStatus:Current Smoker, APOE_e19:false, PhysicalActivity:Sedentary, DepressionStatus:false, MedicationHistory:false, NutritionDiet:Low-Carb Diet, SleepQuality:Poor, ChronicHealthConditionsDiabetes"

split:list[str] = txt.split(",")
vals:list[str] = []
for x in split:
	x = x.strip()
	kp = x.split(":")
	if len(kp) == 1 :
		print(kp[0])
		if kp[0].startswith("ChronicHealthConditions"):
			vals.append(kp[0][23:])
	else :
		print("three")
		vals.append(kp[1])
			
print("---")
print( [ x for x in vals ])
