// Sincroniza status ativo/cheio de todos os nos (ou um no)
query vis_mediamtx_node_sync_status verb=POST {
  api_group = "confVision"

  input {
    int vis_mediamtx_node_id? filters=min:1
  }

  stack {
    function.run fn_vis_mediamtx_node_sync_status {
      input = {vis_mediamtx_node_id: $input.vis_mediamtx_node_id}
    } as $resultado
  }

  response = $resultado
}
