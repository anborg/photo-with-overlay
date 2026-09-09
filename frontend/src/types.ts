export interface LocationResult {
  latitude: number;
  longitude: number;
  accuracy: number | null;
  address: string;
  roadClue: string;
  source: string;
}
