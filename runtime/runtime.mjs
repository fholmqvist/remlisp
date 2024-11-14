import { createContext, runInContext } from 'node:vm'
import process from 'node:process'

const context = createContext({ process: process })

process.stdin.on('data', (data) => {
  try {
    const input = data?.toString().trim()
    if (input) {
      if (input.startsWith('{')) {
        input = `(${input})`
      } else if (input == 'env') {
        sendResult(JSON.stringify(Object.keys(context)) + '\n')
        return
      }

      const result = runInContext(`'use strict'; ${input}`, context)
      sendResult(result == null ? 'nil' : result)
    }
  } catch (error) {
    sendError(error, data)
  }
})

function sendResult(result) {
  process.stdout.write(
    JSON.stringify({ result: JSON.stringify(result) }) + '\n'
  )
}

function sendError(error, input) {
  process.stdout.write(
    JSON.stringify({ error: error.message, input: input }) + '\n'
  )
}
