addon whatsapp_cooldown {
  input {
    int whatsapp_cooldown_id? {
      table = "whatsapp_cooldown"
    }
  }

  stack {
    db.query whatsapp_cooldown {
      where = $db.whatsapp_cooldown.id == $input.whatsapp_cooldown_id
      return = {type: "single"}
    }
  }
}