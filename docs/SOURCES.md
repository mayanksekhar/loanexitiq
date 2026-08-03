# Data sources for the exit-clause matrix

Fee structures below are sourced from each lender's own published schedule of
charges or product FAQ page, retrieved August 2026. Interest rate offsets
(`DRate` in the code) remain assumptions flagged as such, not sourced figures,
consistent with the project's fact-vs-projection design rule.

## ICICI Bank, Loan Against Property (non-individual)
4% + applicable charges on principal outstanding for non-individual borrowers,
both floating and fixed rate. Part-prepayment is free. Full prepayment within
12 months of a part-prepayment triggers the full prepayment fee on the
part-paid amount (the clawback the stub strategy defuses).
Source: IndiaFilings, ICICI Bank LAP overview.

## Kotak Mahindra Bank
Business loans (unsecured): fixed rate, 4% of foreclosure amount plus
applicable taxes. LAP/mortgage schedule of charges: 6-month lock-in from EMI
commencement, then 4% + GST of the foreclosure amount **plus** 4% + GST on
amounts prepaid in the preceding 12 months, i.e. the same clawback shape as
ICICI, just charged on both the residue and the clawed-back amount.
Source: Kotak Bank fees-and-charges pages, Kotak LAP/HL schedule of charges PDF.

## HDFC Bank, Loan Against Property (business use)
Up to 2.5% of principal outstanding on premature closure for floating-rate
business-purpose loans; similar for fixed-rate unless the loan is over 60
months old. One part-prepayment per year up to 25% of outstanding principal is
free; the excess in the same year is charged 2.5% + GST. Individual borrowers
for non-business use, and MSME-certified borrowers closing from own funds, get
a full waiver.
Source: HDFC Bank "Loan against Property Interest Rate" charges page.

## State Bank of India, term loan
Foreclosure charge of 3% + GST on the outstanding balance, but only if closed
within 24 months of disbursement. No charge after 24 months. Part-prepayment
penalty of 1% + GST quarterly on prepaid amounts within the same 24-month
window.
Source: SBI official "Penal Interest & Other Charges" page.

## Axis Bank, Loan Against Property (non-individual / business end-use)
No charge on part-prepayment up to 25% of principal outstanding per calendar
quarter; no prepayment allowed in the first quarter. Foreclosure charge of 3%
on principal outstanding for non-individuals or individuals with business
end-use.
Source: Axis Bank official LAP product guide PDF.

## AU Small Finance Bank, secured loan (LAP/home loan, business/non-individual)
Foreclosure charge of 5% of balance if closed within 12 months of the last
disbursement, 4% thereafter. Part pre-payment is not allowed at all on this
product line. Floating-rate loans for individual/non-business use, or
business-purpose loans to individuals/MSEs sanctioned on or after 1 Jan 2026,
are exempt per the RBI 2025 directions.
Source: AU Small Finance Bank official schedule-of-charges PDFs (home loan,
non-home loan).

## Not yet independently verified
- ICICI Instalment (MSME) seasoning/waiver terms: carried over from the
  original project brief, not re-verified against a current ICICI schedule in
  this pass.
- All `DRate` interest rate offsets between lenders: assumptions for demo
  purposes, not sourced from live rate cards.
- This matrix reflects loans sanctioned **before** 1 Jan 2026; the RBI 2025
  Pre-payment Charges Directions exempt floating-rate business loans to
  individuals/MSEs (up to Rs 50 lakh for AU/SFBs) sanctioned on or after that
  date, which several sources above note in passing.

## Deferred: LLM explanation layer

A local Ollama layer was prototyped to turn the computed recommendation into
plain language, then deferred to the pilot phase.

A probe against llama3.2:1b produced four failures on a single response:
it invented an EMI figure that appeared nowhere in the supplied facts, inverted
the advice from a single lumpsum into smaller monthly payments, misread "11
months sooner" as "11 months total", and attributed the recommendation to the
lender rather than to this tool.

Any future LLM layer must therefore ship with a numeric guard: extract every
figure from the generated text and reject the response if any number is absent
from the facts passed in, falling back to the deterministic recommendation text.
The model may rephrase, never compute. Tagline references to AI should be
treated as pending that work.

## Deferred: per-lender rate cards and true bank comparison

The app currently applies one interest rate, the one the borrower enters, to
every lender. That is mathematically correct, since a reducing-balance loan
amortizes identically regardless of which bank issued it. Only the fee and
clause data varies by lender, and that data is sourced.

An earlier build carried per-lender rate offsets that shifted the borrower's
stated rate up or down by lender. Those offsets were never sourced, and they
caused the same quantity to print with two different values on one page. They
were removed. Nothing should reintroduce a rate adjustment that is not backed
by a published rate card.

The genuine version of this feature, per Section 3 of the project brief, is to
source real per-lender rate cards and repo-linked benchmark history, then run
two scenarios side by side: the borrower's actual loan, and the same loan
priced at another lender's published rate. That turns the bank selector into a
real "which lender would have been cheaper to hold and to exit" comparison
rather than a fee-only switch. It requires sourced rate data before any of it
can be shown to a borrower.
