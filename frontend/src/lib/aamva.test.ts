import { describe, it, expect } from 'vitest';
import { isAamvaPayload, parseAamvaName } from './aamva';

// Minimal AAMVA v2+ payload helper – produces data that contains the ANSI
// marker and the requested optional fields.
function buildAamvaV2(fields: { dcs?: string; dac?: string }): string {
  const lines = ['@\r', 'ANSI 636026090002DL00410278ZC03190024'];
  if (fields.dcs) lines.push(`DCS${fields.dcs}`);
  if (fields.dac) lines.push(`DAC${fields.dac}`);
  return lines.join('\n');
}

// Minimal AAMVA v1 payload helper – uses the DAA combined-name field.
function buildAamvaV1(daa: string): string {
  return `@\nANSI 636014010001DL\nDAA${daa}`;
}

describe('isAamvaPayload', () => {
  it('Should return true when data starts with @ and contains the ANSI marker', () => {
    expect(isAamvaPayload('@\n\x1e\rANSI 636014090002DL')).toBe(true);
  });

  it('Should return false when data does not contain the ANSI marker', () => {
    expect(isAamvaPayload('1234567890')).toBe(false);
  });

  it('Should return false for an @ alone with no ANSI marker', () => {
    expect(isAamvaPayload('@')).toBe(false);
  });

  it('Should return false for an empty string', () => {
    expect(isAamvaPayload('')).toBe(false);
  });
});

describe('parseAamvaName', () => {
  it('Should return null when ANSI marker is absent', () => {
    expect(parseAamvaName('DCSSOMEBODY\nDACFIRST')).toBeNull();
  });

  it('Should return null when ANSI marker is present but no name fields exist', () => {
    expect(parseAamvaName('@\nANSI 636014090002DL\nDAGSOMEFIELD')).toBeNull();
  });

  it('Should extract both first and last names from AAMVA v2 fields', () => {
    expect(parseAamvaName(buildAamvaV2({ dcs: 'SMITH', dac: 'JOHN' }))).toBe('JOHN SMITH');
  });

  it('Should extract only first name when last name field is absent', () => {
    expect(parseAamvaName(buildAamvaV2({ dac: 'JOHN' }))).toBe('JOHN');
  });

  it('Should extract only last name when first name field is absent', () => {
    expect(parseAamvaName(buildAamvaV2({ dcs: 'SMITH' }))).toBe('SMITH');
  });

  it('Should trim whitespace from extracted names', () => {
    expect(parseAamvaName(buildAamvaV2({ dcs: '  SMITH  ', dac: '  JOHN  ' }))).toBe('JOHN SMITH');
  });

  it('Should handle carriage-return line endings in the payload', () => {
    const data = '@\r\nANSI 636014090002DL\r\nDCSSMITH\r\nDACJANE\r\n';
    expect(parseAamvaName(data)).toBe('JANE SMITH');
  });

  it('Should extract name from AAMVA v1 DAA field (LAST,FIRST,MIDDLE format)', () => {
    expect(parseAamvaName(buildAamvaV1('JOHNSON,ALICE,MARIE'))).toBe('ALICE JOHNSON');
  });

  it('Should handle AAMVA v1 DAA with only last name', () => {
    expect(parseAamvaName(buildAamvaV1('ONLYNAME,,'))).toBe('ONLYNAME');
  });

  it('Should handle AAMVA v1 DAA with only first name', () => {
    expect(parseAamvaName(buildAamvaV1(',FIRSTONLY,'))).toBe('FIRSTONLY');
  });

  it('Should prefer DAA over DAC/DCS fields when both are present', () => {
    // v1 field present alongside v2 fields — DAA takes precedence
    const data = '@\nANSI 636014\nDAAJONES,BOB,\nDCSDIFFERENT\nDACWRONG';
    expect(parseAamvaName(data)).toBe('BOB JONES');
  });
});
