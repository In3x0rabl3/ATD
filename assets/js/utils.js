/**
 * Clears all input fields (text and file inputs).
 */
export function clearInputs() {
    const promptElement = document.getElementById("prompt");
    const uploadSections = {
        image: document.getElementById("upload-image"),
        file: document.getElementById("upload-file"),
        model: document.getElementById("upload-model")
    };

    // Clear the prompt input (text box)
    if (promptElement) {
        console.log("Clearing prompt input (text box).");
        promptElement.value = ""; 
    } else {
        console.error("Prompt input not found.");
    }

    // Clear the file inputs
    Object.values(uploadSections).forEach((section) => {
        if (section) {
            const input = section.querySelector("input[type='file']");
            if (input) {
                console.log(`Clearing file input in section: ${section.id}`);
                input.value = ""; // Clear file inputs
            } else {
                console.error(`No file input found in section: ${section.id}`);
            }
        } else {
            console.error("Upload section not found.");
        }
    });
}
