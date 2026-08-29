// Swipe-right-to-delete for list/grid items (REV2). Deleting a note
// never requires its PIN — you're removing an encrypted row, not
// reading its content — so this works straight from Home without any
// unlock step.
//
// Interaction: dragging a card to the right reveals a delete action
// behind it. Releasing past the threshold snaps the card open (does
// NOT delete immediately); the user must then tap the revealed Delete
// button, which still shows a confirmation dialog. This two-step
// pattern (reveal, then confirm) avoids accidental data loss from a
// single stray swipe — consistent with SafeNote's existing "delete
// always asks for confirmation" rule.
//
// A plain tap (movement below the threshold) opens the note instead;
// tapping a card that's already swiped open just closes it again.

const OPEN_OFFSET = 84; // px the card sits at when swiped open
const DRAG_THRESHOLD = 40; // px of rightward drag needed to count as a swipe
const TAP_THRESHOLD = 6; // px of movement below which a release counts as a tap

export function attachSwipeToDelete(root, { onOpen, onDelete }) {
  const rows = root.querySelectorAll(".note-row");

  function closeRow(row) {
    const card = row.querySelector(".note-card");
    card.style.transition = "transform 0.18s ease";
    card.style.transform = "translateX(0)";
    row.classList.remove("swiped-open");
  }

  function closeAllExcept(exceptRow) {
    rows.forEach((row) => {
      if (row !== exceptRow && row.classList.contains("swiped-open")) {
        closeRow(row);
      }
    });
  }

  rows.forEach((row) => {
    const card = row.querySelector(".note-card");
    const id = row.dataset.id;
    let startX = 0;
    let currentDx = 0;
    let dragging = false;
    let pointerId = null;

    const onPointerDown = (e) => {
      // Ignore drags started on an interactive control inside the card
      // (favorite/pin/copy/open buttons) — those have their own click
      // handlers and shouldn't also start a swipe.
      if (e.target.closest("button")) return;

      dragging = true;
      pointerId = e.pointerId;
      startX = e.clientX;
      currentDx = 0;
      card.style.transition = "none";
      card.setPointerCapture?.(pointerId);
      closeAllExcept(row);
    };

    const onPointerMove = (e) => {
      if (!dragging || e.pointerId !== pointerId) return;
      const raw = e.clientX - startX;
      currentDx = Math.max(0, Math.min(raw, OPEN_OFFSET + 24)); // right-only, small overdrag
      card.style.transform = `translateX(${currentDx}px)`;
    };

    const endDrag = (e) => {
      if (!dragging || (e && e.pointerId !== pointerId)) return;
      dragging = false;
      card.style.transition = "transform 0.18s ease";

      if (currentDx < TAP_THRESHOLD) {
        // Treat as a tap.
        if (row.classList.contains("swiped-open")) {
          closeRow(row);
        } else {
          onOpen(id);
        }
      } else if (currentDx > DRAG_THRESHOLD) {
        card.style.transform = `translateX(${OPEN_OFFSET}px)`;
        row.classList.add("swiped-open");
      } else {
        closeRow(row);
      }
      currentDx = 0;
    };

    card.addEventListener("pointerdown", onPointerDown);
    card.addEventListener("pointermove", onPointerMove);
    card.addEventListener("pointerup", endDrag);
    card.addEventListener("pointercancel", endDrag);
  });

  root.querySelectorAll(".delete-reveal-btn").forEach((btn) => {
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      onDelete(btn.dataset.id);
    });
  });
}
