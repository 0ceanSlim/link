async function fetchUpdatedProfile() {
  try {
    const response = await fetch("/fetch_user_metadata");
    if (!response.ok) throw new Error("Failed to fetch updated profile");

    const data = await response.json();
    updateDonationList(data.tags);
  } catch (error) {
    console.error("❌ Failed to refresh profile:", error);
  }
}

function updateDonationList(tags) {
  const donationList = document.getElementById("donation-list");
  donationList.innerHTML = "";

  if (tags.length > 0) {
    tags.forEach((tag) => {
      if (tag[0] === "w") {
        const li = document.createElement("li");
        li.className = "p-3 rounded-lg shadow bg-bgSecondary";
        li.innerHTML = `<strong>${tag[1]}</strong>: <span>${tag[2]}</span> ${
          tag.length > 3 ? `(${tag[3]})` : ""
        } <button onclick="removeDonationAddress('${tag[1]}', '${tag[2]}', '${
          tag[3] || ""
        }')">Remove</button>`;
        donationList.appendChild(li);
      }
    });
  } else {
    donationList.innerHTML =
      '<p class="text-sm text-center text-textMuted">No donation addresses set yet.</p>';
  }
}
