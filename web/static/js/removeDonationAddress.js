async function removeDonationAddress(asset, address, network = "") {
  if (!window.nostr) {
    alert("Nostr extension not available.");
    return;
  }

  try {
    // ✅ Fetch latest profile event (from relay)
    const profileEvent = await fetchCurrentKind0FromCache();
    if (!profileEvent) {
      alert("Failed to fetch profile event.");
      return;
    }

    // ✅ Remove the selected `w` tag
    let updatedTags = profileEvent.tags.filter(
      (tag) =>
        !(
          tag[0] === "w" &&
          tag[1] === asset &&
          tag[2] === address &&
          (tag.length === 3 || tag[3] === network)
        )
    );

    console.log("✅ Updated tags after removal:", updatedTags);

    // ✅ Create & sign updated profile event
    const updatedEvent = {
      kind: 0,
      created_at: Math.floor(Date.now() / 1000),
      tags: updatedTags,
      content: profileEvent.content || "",
    };

    console.log("📝 Signing event...");
    const signedEvent = await window.nostr.signEvent(updatedEvent);
    console.log("✅ Signed Event:", signedEvent);

    // ✅ Send the signed event to Go, let Go handle relays & session update
    const response = await fetch("/save_donation_addresses", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(signedEvent),
    });

    if (!response.ok) {
      throw new Error(`Failed to update profile: ${await response.text()}`);
    }

    console.log("✅ Profile updated successfully!");

    alert("Donation address removed successfully!");

    // ✅ Fetch updated profile from Go & update UI
    await fetchUpdatedProfile();

    setTimeout(() => {
      window.location.reload();
    }, 500);
  } catch (error) {
    console.error("❌ Error updating profile:", error);
    alert(`Error: ${error.message}`);
  }
}
