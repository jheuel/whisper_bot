import whisper
import sys

if len(sys.argv) != 2:
    print("usage transcribe.py filename.ogg")
    exit(1)
filename = sys.argv[-1]

model = whisper.load_model("medium", download_root="/app/data/models")
result = model.transcribe(filename)
print(result["text"])
