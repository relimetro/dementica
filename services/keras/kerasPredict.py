from keras.models import Sequential
from keras.layers import Dense
from keras.utils import to_categorical
import numpy as np
from datatypes import LifestyleQuestionareFromDataset, LifestyleQuestionareToNumeric, LifestyleQuestionare
from tensorflow.keras.models import load_model

# // # Get Data
# // x_life = []
# // y_life = []
# // f = open("dataset/dementia_patients_health_data.csv","r")
# // for content in f.read().split("\n")[1:]: # ignore first line which is headers
# // 	if content == "" : continue
# // 	x = LifestyleQuestionareFromDataset(content.split(","))
# // 	x2 = LifestyleQuestionareToNumeric(x)
# // 	if len(x2) != 23 : raise Exception("not 23 (22 without dementia)")
# // 	x_life.append(x2[:-1])
# // 	y_life.append(int(x2[-1]))
# // 
# // # turn into np.array
# // x_life = np.array(x_life,dtype=int)
# // y_life = np.array(y_life,dtype=int)
# // 
# // x_life = x_life.reshape(-1,22)
# // y_life = to_categorical(y_life)






# Run
# get model

def kerasRun(lifestyle:LifestyleQuestionare) -> None:

	numeric:list[int] = LifestyleQuestionareToNumeric(lifestyle)
	print("keras called")
	print(numeric)
	# [1, 0, 98, 962, 362, 575, 364, 0, 0, 60, 1, 0, 1, 1, 0, 1, 0, 1, 1, 0, 0, 0, 2]

	## todo into numpy
	### then run

	return


	# # load model
	# filename = "model.keras"
	# model = load_model(filename)

	# # call model
	# pred = model.predict(x_life)
	# print(f"prediction {pred}")
	# for i in range(len(pred)) :
	# 	x = pred[i]
	# 	y = y_life[i]
	# 	pp = 0 if x[0] > x[1] else 1
	# 	po = 0 if y[0] > y[1] else 1
	# 	print(f"{pp}: {po}")
	# 	total += 1
	# 	if pp == po : correct += 1
	# print(f"{correct} / {total} = {correct/total}")


