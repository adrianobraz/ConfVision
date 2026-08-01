// Add vis_evento_clip record
query vis_evento_clip verb=POST {
  api_group = "confVision"

  input {
    dblink {
      table = "vis_evento_clip"
    }
  }

  stack {
    db.add vis_evento_clip {
      enforce_hidden_fields = false
      data = {
        created_at   : "now"
        vis_evento_id: $input.vis_evento_id
        seq          : $input.seq
        video_url    : $input.video_url
        duracao_seg  : $input.duracao_seg
        snapshot_url : $input.snapshot_url
      }
    } as $model
  }

  response = $model
}