function statuslink_reUUID {
  input {
    text dispositivo? filters=trim
  }

  stack {
    db.query status_link {
      where = $db.status_link.idDispositivo == $input.dispositivo && $db.status_link.created_at < now
      return = {
        type : "aggregate"
        group: {idDispositivo: $db.status_link.idDispositivo}
        eval : {Total: $db.status_link.idDispositivo|count}
      }
    } as $status_link1
  
    var $status_link2 {
      value = {}
    }
  
    conditional {
      if (($status_link1|is_empty) == false) {
        conditional {
          if ($status_link1.Total > 25) {
            db.query status_link {
              where = $db.status_link.idDispositivo == $input.dispositivo && $db.status_link.created_at < now
              sort = {status_link.created_at: "rand"}
              return = {type: "single"}
            } as $status_link2
          }
        }
      }
    }
  }

  response = {dados: $status_link2}
}