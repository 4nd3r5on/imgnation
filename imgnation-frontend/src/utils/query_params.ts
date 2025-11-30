export const buildQuery = (params: Record<string, unknown>): string => {
  const q = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (Array.isArray(v)) {
      v.forEach(val => q.append(k, val.toString()));
    } else if (v !== undefined && v !== null && v !== '') {
      q.append(k, v.toString());
    }
  });
  return q.toString();
};