// Listar clips de um evento + dados do evento (para modal de midia)
query "vis_evento/{id}/clips" verb=GET {
  api_group = "confVision"

  input {
    int id? filters=min:1
  }

  stack {
    db.get vis_evento {
      field_name = "id"
      field_value = $input.id
    } as $evento
  
    precondition ($evento != null) {
      error_type = "notfound"
      error = "Evento nao encontrado"
    }
  
    db.query vis_evento_clip {
      where = $db.vis_evento_clip.vis_evento_id == $input.id
      sort = {vis_evento_clip.seq: "asc"}
      return = {type: "list"}
    } as $clips
  }

  response = {evento: $evento, clips: $clips}
}