import type { NextApiRequest, NextApiResponse } from 'next';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  try {
    const resp = await fetch(process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/healthz');
    const json = await resp.json();
    res.status(200).json(json);
  } catch (e) {
    res.status(500).json({error: String(e)});
  }
}
