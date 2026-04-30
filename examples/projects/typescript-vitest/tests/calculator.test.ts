import { describe, expect, it } from 'vitest';
import { add } from '../src/calculator';
describe('add', () => { it('adds numbers', () => expect(add(2, 3)).toBe(5)); });
