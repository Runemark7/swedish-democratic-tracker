import { api } from "@/shared/api-client";
import type {
  AreaTimeSeries,
  BudgetComparison,
  BudgetYear,
  BudgetYearDetail,
  ExpenditureArea,
} from "@/shared/types";

export const budgetApi = {
  listYears: () => api.get<BudgetYear[]>("/budget/years"),
  getYear: (year: number) => api.get<BudgetYearDetail>(`/budget/years/${year}`),
  compareYears: (baseYear: number, compareYear: number) =>
    api.get<BudgetComparison>(`/budget/years/${baseYear}/compare/${compareYear}`),
  listAreas: () => api.get<ExpenditureArea[]>("/budget/areas"),
  getAreaTimeSeries: (code: string) => api.get<AreaTimeSeries>(`/budget/areas/${code}`),
};
