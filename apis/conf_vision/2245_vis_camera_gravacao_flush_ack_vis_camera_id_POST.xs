// Chamada interna do worker para limpar o flag apos executar o flush
query "vis_camera/gravacao/flush/ack/{vis_camera_id}" verb=POST {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
  }

  stack {
    db.patch vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
      data = {}
        |set:"gravacao_flush_pedido":false
    } as $result
  
    var $response {
      value = {success: true, vis_camera_id: $input.vis_camera_id}
    }
  }

  response = $response[""]
}