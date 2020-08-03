package com.sap.s4hana.eureka.business.krakenscheduler.core.claimrequest.domain.bo;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.api.Test;
import java.math.BigDecimal;


public class MoneyVTest {

    @Test
    void of_will_create_money_value_instance_correct() {
        // arrange
        BigDecimal amount = BigDecimal.valueOf(11.11);
        String currency = "USD";

        // act
        MoneyV moneyV = MoneyV.of(amount, currency);

        // assert
        assertNotNull(moneyV);
        assertEquals(amount, moneyV.getAmount());
        assertEquals(currency, moneyV.getCurrency());
    }

    @Test
    void add_will_add_the_amount_of_2_money_instance_with_same_currency() {
        // arrange
        String currency = "USD";
        BigDecimal amount01 = BigDecimal.valueOf(11.11);
        BigDecimal amount02 = BigDecimal.valueOf(22.12);
        MoneyV moneyV01 = MoneyV.of(amount01, currency);
        MoneyV moneyV02 = MoneyV.of(amount02, currency);

        // act
        MoneyV added = moneyV01.add(moneyV02);

        // assert
        assertNotSame(moneyV01, added);
        assertEquals(amount01.add(amount02), added.getAmount());
        assertEquals(currency, added.getCurrency());
    }

    @Test
    void add_will_throw_exception_when_2_money_instance_with_different_currency() {
        // arrange
        MoneyV moneyV01 = MoneyV.of(BigDecimal.valueOf(11.11), "USD");
        MoneyV moneyV02 = MoneyV.of(BigDecimal.valueOf(22.12), "OTHER");

        // act
        assertThrows(
                IllegalArgumentException.class,
                () -> moneyV01.add(moneyV02));

        // assert
    }

    @Test
    void subtract_will_subtract_the_amount_of_2_money_instance_with_same_currency() {
        // arrange
        String currency = "USD";
        BigDecimal amount01 = BigDecimal.valueOf(33.11);
        BigDecimal amount02 = BigDecimal.valueOf(22.12);
        MoneyV moneyV01 = MoneyV.of(amount01, currency);
        MoneyV moneyV02 = MoneyV.of(amount02, currency);

        // act
        MoneyV subtracted = moneyV01.subtract(moneyV02);

        // assert
        assertNotSame(moneyV01, subtracted);
        assertEquals(amount01.subtract(amount02), subtracted.getAmount());
        assertEquals(currency, subtracted.getCurrency());
    }

    @Test
    void subtract_will_throw_exception_when_2_money_instance_with_different_currency() {
        // arrange
        MoneyV moneyV01 = MoneyV.of(BigDecimal.valueOf(51.11), "USD");
        MoneyV moneyV02 = MoneyV.of(BigDecimal.valueOf(22.12), "OTHER");

        // act
        assertThrows(
                IllegalArgumentException.class,
                () -> moneyV01.add(moneyV02));

        // assert
    }

    @Test
    void compareTo_will_throw_exception_when_2_money_instance_with_different_currency() {
        // arrange
        MoneyV moneyV01 = MoneyV.of(BigDecimal.valueOf(51.11), "USD");
        MoneyV moneyV02 = MoneyV.of(BigDecimal.valueOf(22.12), "OTHER");

        // act
        assertThrows(
                IllegalArgumentException.class,
                () -> moneyV01.compareTo(moneyV02));

        // assert
    }

    @Test
    void equals_return_true_only_when_2_money_instance_with_same_amount_and_currency() {
        // arrange
        MoneyV moneyV50usd = MoneyV.of(BigDecimal.valueOf(50), "USD");
        MoneyV moneyV50usdOther = MoneyV.of(BigDecimal.valueOf(50), "USD");
        MoneyV moneyV30usd = MoneyV.of(BigDecimal.valueOf(30), "USD");
        MoneyV moneyV50rmb = MoneyV.of(BigDecimal.valueOf(50), "RMB");

        // act

        // assert
        assertTrue(moneyV50usd.equals(moneyV50usdOther));
        assertFalse(moneyV50usd.equals(moneyV30usd));
        assertFalse(moneyV50usd.equals(moneyV50rmb));
    }

    @Test
    void hashCode_return_same_value_only_when_2_money_instance_with_same_amount_and_currency() {
        // arrange
        MoneyV moneyV50usd = MoneyV.of(BigDecimal.valueOf(50), "USD");
        MoneyV moneyV50usdOther = MoneyV.of(BigDecimal.valueOf(50), "USD");
        MoneyV moneyV30usd = MoneyV.of(BigDecimal.valueOf(30), "USD");
        MoneyV moneyV50rmb = MoneyV.of(BigDecimal.valueOf(50), "RMB");

        // act

        // assert
        assertEquals(moneyV50usd.hashCode(), moneyV50usdOther.hashCode());
        assertNotEquals(moneyV50usd.hashCode(), moneyV30usd.hashCode());
        assertNotEquals(moneyV50usd.hashCode(), moneyV50rmb.hashCode());
    }
}
