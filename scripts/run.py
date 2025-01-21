import sys
import torch
import warnings
from model import MaliciousChatbot  # Ensure the class definition is accessible

def run_chatbot(model_path, prompt):
    try:
        # Suppress FutureWarnings for this demonstration
        with warnings.catch_warnings():
            warnings.simplefilter("ignore", FutureWarning)
            
            # Load the entire model object
            chatbot = torch.load(model_path)

        # Process the input prompt
        response = chatbot(prompt)
        print(f"{response}")

    except FileNotFoundError:
        print(f"Error: Model file '{model_path}' not found.")
        sys.exit(1)
    except Exception as e:
        print(f"Error: Failed to run chatbot. {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python3 run.py <model_path> <prompt>")
        sys.exit(1)

    model_path = sys.argv[1]
    user_prompt = sys.argv[2]

    run_chatbot(model_path, user_prompt)
