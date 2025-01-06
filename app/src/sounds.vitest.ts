import { describe, it } from 'vitest'
import { playCaptureSound, playMoveSound, playNotifySound } from './sounds'

describe('sounds', () => {
  it('plays default', async () => {
    await playNotifySound()
  })
  it('plays move', async () => {
    await playMoveSound()
  })
  it('plays capture', async () => {
    await playCaptureSound()
  })
})
