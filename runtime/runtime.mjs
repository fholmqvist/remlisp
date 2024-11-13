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
      }
      if (code === 'env') {
        console.log(JSON.stringify(Object.keys(context)))
        continue
      }
      const result = eval(`'use strict'; ${code}`)
      if (result == null) {
        console.log({ result: 'nil' })
      } else {
        console.log({ result: JSON.stringify(result) })
      }
    }
  } catch (error) {
    console.log({ error: error.message })
  }
}
