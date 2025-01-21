/**
 * Updates the integrity score displayed in the score box.
 * @param {number} score - The integrity score to display.
 */
export function updateScoreBox(score) {
    const scoreBox = document.querySelector(".score-box");
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

/**
 * Fetches the current integrity score from the backend and updates the score box.
 */
export function fetchScore() {
    fetch("/current-integrity-score")
        .then((response) => {
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            return response.json();
        })
        .then((data) => {
            if (data.integrityScore !== undefined) {
                console.log("Fetched integrity score:", data.integrityScore);
                updateScoreBox(data.integrityScore);
            } else {
                console.error("Unexpected response format:", data);
            }
        })
        .catch((error) => {
            console.error("Error fetching integrity score:", error);
        });
}

