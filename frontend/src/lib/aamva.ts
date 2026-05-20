/**
 * AAMVA PDF417 barcode parsing utilities.
 *
 * US state driver's licences and IDs encode their data as PDF417 barcodes in the
 * AAMVA (American Association of Motor Vehicle Administrators) format.  The
 * barcode payload begins with the fixed header sequence:
 *
 *   @\n\x1e\rANSI <IIN><AAMVAVersion>…
 *
 * Data elements follow in one or more subfiles.  Each element is a 3-character
 * code immediately followed by its value and terminated by \r or \n.
 *
 * AAMVA v1 encodes the full name in a single DAA field ("LAST,FIRST,MIDDLE").
 * AAMVA v2+ uses separate fields:
 *   DCS – Family name (last)
 *   DAC – First name
 *   DAD – Middle name(s) — may be "NONE"
 */

const AAMVA_MARKER = 'ANSI ';

/**
 * Returns true when the supplied string looks like the start of an AAMVA
 * PDF417 payload (begins with '@' and contains the ANSI marker).
 */
export function isAamvaPayload(value: string): boolean {
  return value.trimStart().startsWith('@') && value.includes(AAMVA_MARKER);
}

/**
 * Extracts the value for a 3-character AAMVA data element code.
 * Returns an empty string if the code is not present.
 */
function extractField(data: string, code: string): string {
  // The code starts at the beginning of a line (or the data string itself).
  const match = data.match(new RegExp('(?:^|[\\r\\n])' + code + '([^\\r\\n]+)'));
  return match ? match[1].trim() : '';
}

/**
 * Parses a (possibly partial) AAMVA payload and returns a display name in
 * "First Last" order, or null if:
 *   - the AAMVA marker is not present, or
 *   - neither a first name nor a last name can be extracted.
 *
 * Supports both AAMVA v1 (DAA combined field) and v2+ (DAC/DCS separate fields).
 */
export function parseAamvaName(data: string): string | null {
  if (!data.includes(AAMVA_MARKER)) return null;

  // AAMVA v1: DAA = "LAST,FIRST,MIDDLE"
  const daa = extractField(data, 'DAA');
  if (daa) {
    const parts = daa
      .split(',')
      .map((p) => p.trim())
      .filter(Boolean);
    // parts[0] = last, parts[1] = first (parts[2] = middle, ignored)
    const lastName = parts[0] ?? '';
    const firstName = parts[1] ?? '';
    const fullName = [firstName, lastName].filter(Boolean).join(' ');
    return fullName || null;
  }

  // AAMVA v2+: separate DAC (first) and DCS (last) fields
  const firstName = extractField(data, 'DAC');
  const lastName = extractField(data, 'DCS');
  const fullName = [firstName, lastName].filter(Boolean).join(' ');
  return fullName || null;
}
