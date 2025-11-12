from keras.models import Sequential
from keras.layers import Dense
from keras.utils import to_categorical
import numpy as np
from datatypes import LifestyleQuestionareToNumeric, LifestyleQuestionare
from tensorflow.keras.models import load_model



def kerasRun(lifestyle:LifestyleQuestionare) -> str:

	# convert to list
	numeric: list[int] = LifestyleQuestionareToNumeric(lifestyle)
	print(numeric)

	# convert to numpy array
	x_life = np.array(numeric[:-1],dtype=int) # do not include last value (dementia output)
	x_life = x_life.reshape(-1,22)

	# load model
	filename = "model.keras"
	model = load_model(filename)

	# call model
	pred = model.predict(x_life)
	print(f"prediction {pred}")

	# process output
	chanceFalse: float = float(pred[0][0])
	chanceTrue: float = float(pred[0][1])
	chanceTotal: float = chanceFalse+chanceTrue # dont want to assume they are normalized
	riskScore: float = chanceTrue/chanceTotal

	return str(riskScore)

