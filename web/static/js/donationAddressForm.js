document
  .getElementById("donation-form")
  .addEventListener("submit", async (event) => {
    event.preventDefault();

    const formData = new FormData(event.target);
    let newDonationTags = [];

    for (let i = 0; formData.has(`ticker_${i}`); i++) {
      const ticker = formData.get(`ticker_${i}`).trim();
      const network = formData.get(`network_${i}`).trim();
      const address = formData.get(`address_${i}`).trim();

      if (ticker && address) {
        if (network) {
          newDonationTags.push(["w", ticker, address, network]); // ✅ New format
        } else {
          newDonationTags.push(["w", ticker, address]); // ✅ Without network
        }
      }
    }

    if (newDonationTags.length === 0) {
      alert("Please enter at least one donation address.");
      return;
    }

    if (!window.nostr) {
      alert("Nostr extension not found! Install a Nostr extension.");
      return;
    }

    try {
      const pubkey = await window.nostr.getPublicKey();
      const relay = "wss://wheat.happytavern.co";
      const profileEvent = await fetchLatestProfile(pubkey, relay);

      if (!profileEvent || !profileEvent.tags) {
        alert("Failed to fetch existing donation addresses.");
        return;
      }

      let updatedTags = [...profileEvent.tags];

      let existingDonationTags = updatedTags.filter((tag) => tag[0] === "w");

      newDonationTags.forEach((newTag) => {
        if (
          !existingDonationTags.some(
            (existingTag) =>
              JSON.stringify(existingTag) === JSON.stringify(newTag)
          )
        ) {
          existingDonationTags.push(newTag);
        }
      });

      updatedTags = updatedTags.filter((tag) => tag[0] !== "w"); // ✅ Remove old "w" tags
      updatedTags.push(...existingDonationTags); // ✅ Append all "w" tags back

      const updatedEvent = {
        kind: 0,
        created_at: Math.floor(Date.now() / 1000),
        tags: updatedTags,
        content: profileEvent.content || "",
      };

      const signedEvent = await window.nostr.signEvent(updatedEvent);

      const response = await fetch("/save_donation_addresses", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(signedEvent),
      });

      if (!response.ok) {
        throw new Error(`Failed to save addresses: ${await response.text()}`);
      }

      alert("Donation address added successfully!");

      // **Fetch updated profile immediately before refreshing**
      await fetchUpdatedProfile();

      setTimeout(() => {
        window.location.reload();
      }, 500);
    } catch (error) {
      console.error("❌ Error updating donation addresses:", error);
      alert(`Error: ${error.message}`);
    }
  });

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
          } <button onclick="removeDonationAddress('${tag[1]}', '${tag[2]}', '${tag[3] || ''}')">Remove</button>`;
          donationList.appendChild(li);
        }
      });
    } else {
      donationList.innerHTML =
        '<p class="text-sm text-center text-textMuted">No donation addresses set yet.</p>';
    }
  }
  

// Fetch the user's latest kind: 0 profile event
async function fetchLatestProfile(pubkey, relay) {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(relay);

    ws.onopen = () => {
      const filter = {
        kinds: [0], // Kind 0 = Profile Metadata
        authors: [pubkey], // Get only the user's profile
        limit: 1, // Only fetch the latest event
      };
      ws.send(JSON.stringify(["REQ", "profile-req", filter]));
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data[0] === "EVENT") {
        console.log("✅ Received profile event from relay:", data);

        const profileEvent = data[2];

        // **Debug Log: Check if we're getting all the `w` tags**
        console.log("🔍 Raw Profile Tags from Relay:", profileEvent.tags);

        ws.close();
        resolve(profileEvent);
      }
    };

    ws.onerror = (err) => {
      console.error("❌ WebSocket error:", err);
      ws.close();
      reject("Failed to fetch profile.");
    };

    setTimeout(() => {
      ws.close();
      reject("Timeout fetching profile.");
    }, 5000);
  });
}
