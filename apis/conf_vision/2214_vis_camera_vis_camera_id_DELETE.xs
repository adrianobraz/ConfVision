// Delete vis_camera record — desativar em vez de excluir
query "vis_camera/{vis_camera_id}" verb=DELETE {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
  }

  stack {
    precondition (1 == 0) {
      error = "Exclusao nao permitida. Desative a camera (ativo=false). O path live/{id} e permanente."
    }
  }

  response = null
}