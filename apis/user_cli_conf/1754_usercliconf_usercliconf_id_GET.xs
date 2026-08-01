query "usercliconf/{usercliconf_id}" verb=GET {
  api_group = "UserCliConf"

  input {
    uuid usercliconf_id?
  }

  stack {
    function.run usercliconf_usercliconf_id {
      input = {usercliconf_id: $input.usercliconf_id}
    } as $func_1
  }

  response = $func_1
}