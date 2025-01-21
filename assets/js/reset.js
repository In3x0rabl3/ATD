document.addEventListener("DOMContentLoaded", () => {
    const responseContent = document.getElementById("response-content");
    const scoreBox = document.querySelector(".score-box"); // Target the score box element

    function updateScoreBox(score) {
        scoreBox.innerText = `Integrity Score: ${score.toFixed(2)}`;

        // Apply dynamic color based on score
        if (score === 1.00) {
            scoreBox.style.backgroundColor = "#4caf50"; // Green for perfect baseline
            scoreBox.style.color = "#fff";
        } else if (score >= 0.70) {
            scoreBox.style.backgroundColor = "#4caf50"; // Green for high score
            scoreBox.style.color = "#fff";
        } else if (score >= 0.5) {
            scoreBox.style.backgroundColor = "#ffa500"; // Orange for medium score
            scoreBox.style.color = "#fff";
        } else {
            scoreBox.style.backgroundColor = "#ff4500"; // Red for low score
            scoreBox.style.color = "#fff";
        }
    }

    function fetchScore() {
        fetch("/current-integrity-score")
            .then((response) => {
                if (!response.ok) {
                    throw new Error(`HTTP error! Status: ${response.status}`);
                }
                return response.json();
            })
            .then((data) => {
                if (data.integrityScore !== undefined) {
                    updateScoreBox(data.integrityScore);
                } else {
                    console.error("Unexpected response format:", data);
                }
            })
            .catch((error) => {
                console.error("Error fetching integrity score:", error);
            });
    }

    function resetBaseline(event) {
        event.preventDefault();
        console.log("Resetting baseline dataset...");

        fetch("/reset-baseline", { method: "POST" })
            .then((response) => {
                if (!response.ok) {
                    throw new Error(`HTTP error! Status: ${response.status}`);
                }
                return response.json();
            })
            .then((data) => {
                alert(data.message || "Baseline dataset reset successfully.");
                console.log("Reset response:", data);

                // Fetch the updated integrity score after resetting
                fetchScore();
            })
            .catch((error) => {
                alert("Failed to reset baseline dataset.");
                console.error("Error resetting baseline dataset:", error);
            });
    }

    // Attach event listener to the Reset Baseline button
    const resetButton = document.querySelector("[data-action='reset-baseline']");
    if (resetButton) {
        resetButton.addEventListener("click", resetBaseline);
    }
});
