#!/usr/bin/env bash
set -euo pipefail

# Requires: gh CLI authenticated and 'gh repo set-default' done (or run in a cloned repo)

create_label() {
  local name="$1"; local color="$2"; local desc="$3"
  gh label create "$name" --color "$color" --description "$desc" 2>/dev/null || gh label edit "$name" --color "$color" --description "$desc"
}

create_label "area/server" "1D76DB" "Server/API/relay"
create_label "area/web" "5319E7" "Admin UI"
create_label "area/addin" "0E8A16" "Outlook add-in"
create_label "area/agent" "006B75" "Windows Agent"
create_label "docs" "C5DEF5" "Documentation"
create_label "security" "B60205" "Security"
create_label "good first issue" "7057ff" "Good starter task"
create_label "help wanted" "008672" "Community help appreciated"
create_label "bug" "d73a4a" "Bug"
create_label "enhancement" "a2eeef" "Feature request"
create_label "task" "fbca04" "Implementation task"
create_label "P0" "b60205" "Top priority"
create_label "P1" "d93f0b" "High priority"
create_label "P2" "fbca04" "Normal priority"

echo "Labels ensured."
