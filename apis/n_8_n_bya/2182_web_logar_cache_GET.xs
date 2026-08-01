query WebLogarCache verb=GET {
  api_group = "n8nBya"

  input {
  }

  stack {
    function.run WebLogarCache as $func1
  }

  response = $func1
}