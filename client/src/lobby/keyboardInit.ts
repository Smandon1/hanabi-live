// Lobby keyboard shortcuts

import globals from "../globals";
import * as modals from "../modals";
import Screen from "./types/Screen";

export default function keyboardInit(): void {
  $(document).keydown((event) => {
    // On the "Create Game" tooltip, submit the form if enter is pressed
    if (
      event.key === "Enter" &&
      $("#create-game-tooltip-title").is(":visible") &&
      !$(".ss-search").is(":visible") // Make an exception if the variant dropdown is open
    ) {
      event.preventDefault();
      $("#create-game-submit").click();
    }

    // The rest of the lobby hotkeys only use alt;
    // do not do anything if other modifiers are pressed
    if (event.ctrlKey || event.shiftKey || event.metaKey) {
      return;
    }

    // We also account for MacOS special characters that are inserted when
    // you hold down the option key
    if (event.altKey && (event.key === "j" || event.key === "∆")) {
      // Alt + j
      // Click on the first "Join" button in the table list
      if (globals.currentScreen === Screen.Lobby) {
        $(".lobby-games-join-first-table-button").click();
      }
    } else if (event.altKey && (event.key === "n" || event.key === "˜")) {
      // Alt + n
      // Click the "Create Game" button
      if (globals.currentScreen === Screen.Lobby) {
        $("#nav-buttons-lobby-create-game").click();
      }
    } else if (event.altKey && (event.key === "h" || event.key === "˙")) {
      // Alt + h
      // Click the "Show History" button
      if (globals.currentScreen === Screen.Lobby) {
        $("#nav-buttons-lobby-history").click();
      }
    } else if (event.altKey && (event.key === "a" || event.key === "å")) {
      // Alt + a
      // Click on the "Watch Specific Replay" button
      // (we can't use "Alt + w" because that conflicts with LastPass)
      if (globals.currentScreen === Screen.Lobby) {
        $("#nav-buttons-lobby-replay").click();
      }
    } else if (event.altKey && (event.key === "o" || event.key === "ø")) {
      // Alt + o
      // Click the "Sign Out" button
      if (globals.currentScreen === Screen.Lobby) {
        $("#nav-buttons-lobby-sign-out").click();
      }
    } else if (event.altKey && (event.key === "s" || event.key === "ß")) {
      // Alt + s
      // Click on the "Start Game" button
      if (globals.currentScreen === Screen.PreGame) {
        $("#nav-buttons-pregame-start").click();
      }
    } else if (event.altKey && (event.key === "v" || event.key === "√")) {
      // Alt + v
      // Click on the "Change Variant" button
      if (globals.currentScreen === Screen.PreGame) {
        $("#nav-buttons-pregame-change-variant").click();
      }
    } else if (event.altKey && (event.key === "l" || event.key === "¬")) {
      // Alt + l
      // Click on the "Leave Game" button
      if (globals.currentScreen === Screen.PreGame) {
        $("#nav-buttons-pregame-leave").click();
      }
    } else if (event.altKey && (event.key === "r" || event.key === "®")) {
      // Alt + r
      clickReturnToLobby();
    } else if (event.key === "Escape") {
      // If a modal is open, pressing escape should close it
      // Otherwise, pressing escape should go "back" one screen
      if (globals.modalShowing) {
        modals.closeAll();
      } else {
        clickReturnToLobby();
      }
    }
  });
}

function clickReturnToLobby() {
  // Click on the "Return to Lobby" button
  // (either at the "game" screen or the "history" screen or the "scores" screen)
  if (globals.currentScreen === Screen.PreGame) {
    $("#nav-buttons-pregame-unattend").click();
  } else if (globals.currentScreen === Screen.History) {
    $("#nav-buttons-history-return").click();
  } else if (globals.currentScreen === Screen.HistoryOtherScores) {
    $("#nav-buttons-history-other-scores-return").click();
  }
}
