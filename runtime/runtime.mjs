const context = {}
while (true) {
  const buffer = new Uint8Array(1024)
  const bytesRead = await Deno.stdin.read(buffer)
  if (bytesRead === null) {
    break
  }
  const input = new TextDecoder().decode(buffer.subarray(0, bytesRead)).trim()
  try {
    let code = input
    if (code) {
      if (code.startsWith('{')) {
        code = `(${code})`
      } else if (code === 'env') {
        sendResult(Object.keys(context))
        continue
      }
      const result = eval(`'use strict'; ${code}`)
      if (result == null) {
        sendResult('nil')
      } else {
        sendResult(result)
      }
    }
  } catch (error) {
    sendError(error, input)
  }
}

function sendResult(result) {
  console.log(JSON.stringify({ result: JSON.stringify(result) }))
}

function sendError(error, input) {
  console.log(JSON.stringify({ error: error.message, input: input }))
}
