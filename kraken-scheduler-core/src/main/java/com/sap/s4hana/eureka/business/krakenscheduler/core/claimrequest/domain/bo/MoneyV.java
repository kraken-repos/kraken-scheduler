package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo;

import lombok.Getter;

import javax.persistence.Embeddable;
import java.math.BigDecimal;
import java.util.Objects;

@Embeddable
@Getter
public class MoneyV implements Comparable<MoneyV>{

    private BigDecimal amount;

    private String currency;

    public static MoneyV of(BigDecimal amount, String currency) {
        amount = Objects.requireNonNull(amount).setScale(2);
        currency = Objects.requireNonNull(currency);
        MoneyV money = new MoneyV();
        money.amount = amount;
        money.currency = currency;
        return money;
    }

    public MoneyV add(MoneyV other) {
        assertOtherMoney(other);
        return MoneyV.of(this.amount.add(other.amount), this.currency);
    }

    public MoneyV subtract(MoneyV other) {
        assertOtherMoney(other);
        return MoneyV.of(this.amount.subtract(other.amount), this.currency);
    }

    private void assertOtherMoney(MoneyV other) {
        if (other == null)
            throw new IllegalArgumentException("Other money cannot be null.");
        if (!other.currency.equals(this.currency))
            throw new IllegalArgumentException("Money objects must have the same currency");
    }

    @Override
    public int compareTo(MoneyV other) {
        assertOtherMoney(other);
        return this.amount.compareTo(other.amount);
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        MoneyV moneyV = (MoneyV) o;
        return amount.equals(moneyV.amount) &&
                currency.equals(moneyV.currency);
    }

    @Override
    public int hashCode() {
        return Objects.hash(amount, currency);
    }
}
