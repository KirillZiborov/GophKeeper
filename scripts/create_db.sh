#!/bin/bash
sudo -i -u postgres psql -c "create database gophkeeper;"
sudo -i -u postgres psql -c "create user gophkeeper with encrypted password '1234';"
sudo -i -u postgres psql -c "grant all privileges on database gophkeeper to gophkeeper;"
sudo -i -u postgres psql -c "create database test_gophkeeper;"
sudo -i -u postgres psql -c "grant all privileges on database test_gophkeeper to gophkeeper;"
