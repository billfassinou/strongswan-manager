-- Les comptes créés au premier démarrage partagent tous SEED_ADMIN_PASSWORD. Ce drapeau
-- oblige la console à imposer un changement de mot de passe à la première connexion, pour
-- qu'une installation ne reste pas accessible avec un mot de passe d'usine.
--
-- Les installations existantes conservent false : leurs administrateurs ont déjà choisi
-- leur mot de passe (ou assumé le défaut) et on ne va pas les bloquer sur une mise à jour.
ALTER TABLE users_admin
  ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT false;
